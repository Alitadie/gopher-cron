package main

import (
	"context"
	"fmt"
	"gopher-cron/internal/core"
	"gopher-cron/internal/infra"
	"log"
	"time"
)

// 为了演示，我们将 Worker 端口定死。实际需动态分配。
const WorkerPort = "9090"

func main() {
	// --- 组件初始化 ---
	jobManager := core.NewJobManager()

	// 1. 初始化核心任务（供 Worker 使用）
	jobID := jobManager.AddJob("test-job", func(ctx context.Context) error {
		fmt.Println("🚀 WORKER: executing actual job logic...")
		return nil
	})

	// 启动 JobManager 内部消费协程
	go jobManager.Start(context.Background())

	// 2. 启动 Worker 的 gRPC Server
	grpcWorker := infra.NewGrpcWorker(WorkerPort, jobManager)
	go func() {
		if err := grpcWorker.Start(); err != nil {
			panic(err)
		}
	}()

	// 给予 Server 启动缓冲时间
	time.Sleep(1 * time.Second)

	// --- 模拟 Master 行为 ---
	// 3. 模拟 Master 决定调度任务
	log.Println("MASTER: Starting dispatch sequence...")
	dispatcher := infra.NewMasterDispatcher()

	// 模拟从 Etcd 拿到的 Worker 地址 (localhost:9090)
	targetWorker := "localhost:" + WorkerPort

	err := dispatcher.Dispatch(context.Background(), targetWorker, jobID, "Testing")
	if err != nil {
		log.Printf("MASTER: Dispatch failed: %v", err)
	} else {
		log.Println("MASTER: Dispatch Success! ✅")
	}

	// 阻塞住，看效果
	select {}
}
