package core_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gopher-cron/internal/core"
)

// BDD: Given 一个任务管理器, When 我添加一个每 100ms 执行一次的任务, Then 1秒后计数器应该大约增加 10 次
func TestScheduler_HighConcurrencyExecution(t *testing.T) {
	manager := core.NewJobManager()

	var counter int32
	// 模拟一个具体任务
	jobID := manager.AddJob("TestJob", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		t.Logf("Worker running: %d", counter)
		return nil
	})

	// 启动调度器（非阻塞）
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go manager.Start(ctx) // 核心：启动调度循环

	// 模拟每 100ms 触发一次任务（在真实系统中，这里是解析 Cron 表达式）
	// 为了 MVP，我们手动模拟触发 5 次并发执行
	for i := 0; i < 5; i++ {
		go manager.TriggerJob(jobID) // 手动触发
		time.Sleep(100 * time.Millisecond)
	}

	// 等待上下文结束
	<-ctx.Done()

	finalCount := atomic.LoadInt32(&counter)
	if finalCount != 5 {
		t.Errorf("期望执行 5 次, 实际执行 %d 次", finalCount)
	}
}
