package infra_test

import (
	"context"
	"gopher-cron/internal/infra"
	"testing"
	"time"
)

// 这里我们需要真实的 Etcd 环境，建议先跑 docker
// TestElection 模拟两个节点选主
func TestElection(t *testing.T) {
	// 模拟节点 1
	c1, err := infra.NewCoordinator("node-1", []string{"localhost:2379"})
	if err != nil {
		t.Skip("Skipping test: Etcd might not be running")
		return
	}

	// 模拟节点 2
	c2, _ := infra.NewCoordinator("node-2", []string{"localhost:2379"})

	// 结果通道
	leaderCh := make(chan string, 2)

	// 两个节点同时竞选
	go func() {
		if c1.Campaign(context.Background()) {
			leaderCh <- "node-1"
		}
	}()

	go func() {
		//稍微慢一点启动，避免绝对冲突难以观察
		time.Sleep(50 * time.Millisecond)
		if c2.Campaign(context.Background()) {
			leaderCh <- "node-2"
		}
	}()

	// 设定时间观测谁成为了 Leader
	select {
	case leader := <-leaderCh:
		t.Logf("Master elected: %s", leader)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: No leader elected")
	}

	// 清理资源（关闭 Session 会自动释放锁）
	c1.Close()
	c2.Close()
}
