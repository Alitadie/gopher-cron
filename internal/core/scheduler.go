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
	rwMuxtex sync.RWMutex //读写锁
	jobQueue chan string  //缓冲 充当队列
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs:     make(map[string]*Job),
		jobQueue: make(chan string, 100), //缓冲100个 防止任务瞬时 暴增阻塞
	}
}

// 1.注册任务 写锁
func (m *JobManager) AddJob(name string, logic JobLogic) string {
	m.rwMuxtex.Lock()
	defer m.rwMuxtex.Unlock()

	id := fmt.Sprintf("job-%s", name)
	m.jobs[id] = &Job{
		ID:    id,
		Name:  name,
		Logic: logic,
	}

	return id
}

// 2. 触发任务 不直接 执行 丢进channel
func (m *JobManager) TriggerJob(jobID string) {
	// 非阻塞 满了丢弃或报错
	select {
	case m.jobQueue <- jobID:
		//成功入队
	default:
		fmt.Printf("Warning: Job Queue full, drop job %s\n", jobID)
	}
}

// 3. 启动调度器 消费者模式
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
			//优雅退出
			fmt.Println("Scheduler Stopped...")
			return
		}
	}
}

func (m *JobManager) executeWorker(ctx context.Context, jobID string) {
	//读锁 获取任务详情
	m.rwMuxtex.RLock()
	job, exists := m.jobs[jobID]
	m.rwMuxtex.RUnlock()

	if !exists {
		fmt.Printf("Warning: Job %s not found\n", jobID)
		return
	}
	// 核心知识点：Panic Recover (微服务一定要防崩)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Job %s PANIC: %v\n", job.Name, r)
		}
	}()

	// 2. 执行 Job
	fmt.Printf("Job %s started...\n", jobID)
	if err := job.Logic(ctx); err != nil {
		fmt.Printf("Error: Job %s failed: %v\n", jobID, err)
	}
	fmt.Printf("Job %s finished...\n", jobID)
}
