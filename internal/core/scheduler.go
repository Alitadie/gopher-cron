package core

import (
	"context"
	"fmt"
	"sync"
)

// 领域对象 Job
type JobLogic func(ctx context.Context) error

type Job struct {
	ID    string
	Name  string
	Logic JobLogic
}

// 聚合根 JobManager
type JobManager struct {
	jobs     map[string]*Job
	rwMutex  sync.RWMutex // 读写锁：保护 map，防止并发读写 panic
	jobQueue chan string  // 缓冲 Channel：充当异步任务队列
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs:     make(map[string]*Job),
		jobQueue: make(chan string, 100), // 缓冲 100 个，防止任务瞬时暴增阻塞调度器
	}
}

// 1. 注册任务 (写锁)
func (m *JobManager) AddJob(name string, logic JobLogic) string {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()

	id := fmt.Sprintf("job-%s", name) // 简化生成 ID
	m.jobs[id] = &Job{
		ID:    id,
		Name:  name,
		Logic: logic,
	}
	return id
}

// 2. 触发任务 (并不直接执行，而是丢进 Channel)
func (m *JobManager) TriggerJob(jobID string) {
	// 非阻塞式发送，如果队列满了，选择丢弃或报错（高并发系统的自我保护）
	select {
	case m.jobQueue <- jobID:
		// 成功入队
	default:
		fmt.Printf("Warning: Job Queue full, drop job %s\n", jobID)
	}
}

// 3. 启动调度引擎 (消费者模式)
func (m *JobManager) Start(ctx context.Context) {
	fmt.Println("Scheduler Started...")

	// 启动一定数量的 Worker 协程（这里简单起见，来一个起一个，或使用 Worker Pool）
	// 本次 MVP 采用：主循环监听 channel，收到任务 ID 此时起 Goroutine 执行
	for {
		select {
		case jobID := <-m.jobQueue:
			// 核心考点：Go 的闭包陷阱。如果这里不用传参，可能 jobID 变量会错乱
			// 方案：开启新的 Goroutine 执行任务，不阻塞调度循环
			go m.executeWorker(ctx, jobID)

		case <-ctx.Done():
			// 优雅退出
			fmt.Println("Scheduler Stopped.")
			return
		}
	}
}

func (m *JobManager) executeWorker(ctx context.Context, jobID string) {
	// 读锁：获取任务详情
	m.rwMutex.RLock()
	job, exists := m.jobs[jobID]
	m.rwMutex.RUnlock()

	if !exists {
		return
	}

	// 核心知识点：Panic Recover (微服务一定要防崩)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Job %s PANIC: %v\n", job.Name, r)
		}
	}()

	// 执行逻辑
	fmt.Printf("[%s] executing...\n", job.Name)
	_ = job.Logic(ctx)
}
