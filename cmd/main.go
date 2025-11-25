package main

import (
	"context"
	"gopher-cron/internal/core"
	"gopher-cron/internal/infra"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 模拟从环境变量获取 Node ID
	nodeID := "node-" + os.Getenv("HOSTNAME")
	if nodeID == "node-" {
		nodeID = "node-local-1"
	}

	// 1. 连接 Etcd
	coord, err := infra.NewCoordinator(nodeID, []string{"127.0.0.1:2379"})
	if err != nil {
		log.Fatalf("Failed to connect etcd: %v", err)
	}
	defer coord.Close()

	// 2. 注册节点 (我是个干活的 Worker)
	go func() {
		if err := coord.RegisterNode(); err != nil {
			log.Printf("Register error: %v", err)
		}
	}()

	// 3. 初始化调度核心
	jobManager := core.NewJobManager()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. 核心分歧点：是 Master 才跑调度
	// 开启一个协程去竞选
	go func() {
		// 如果竞选成功，代码才会往下走；如果别人是 Master，这里会阻塞住等待机会
		if coord.Campaign(ctx) {
			// --- 我是 Master ---
			log.Println("Main Scheduler starting...")
			// 这里我们简单模拟加入一个任务
			jobManager.AddJob("DistributedJob", func(c context.Context) error {
				log.Println("Execute Logic...")
				return nil
			})
			// 启动调度循环
			jobManager.Start(ctx)
		} else {
			// Campaign 返回 false 说明 ctx 可能被取消了，或者发生了错误
			log.Println("Campaign stopped")
		}
	}()

	// 5. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	cancel() // 这会导致 Campaign context 取消，Master 放弃身份
}
