package core_test

import (
	"context"
	"gopher-cron/internal/core"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_HighConcurrencyExecution(t *testing.T) {
	manager := core.NewJobManager()

	var counter int32

	//模拟一个具体任务
	jobID := manager.AddJob("TestJob", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		t.Logf("Worker running: %d", counter)
		return nil
	})
	//启动一个调度器 非阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go manager.Start(ctx) //启动调度器

	// 模拟每100ms触发一次任务 真实系统中是解析 cron表达式
	// 手动触发5次并发执行
	for i := 0; i < 5; i++ {
		go manager.TriggerJob(jobID) //手动触发
		time.Sleep(100 * time.Millisecond)
	}

	// 等待调度器完成
	<-ctx.Done()

	finalCount := atomic.LoadInt32(&counter)
	if finalCount != 5 {
		t.Errorf("Expected counter to be 5, but got %d", finalCount)
	}
}
