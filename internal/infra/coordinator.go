package infra

import (
	"context"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	ElectionKey = "/gopher-cron/election"
	RegistryKey = "/gopher-cron/workers/"
	TTL         = 10 // 租约时间 10s
)

type Coordinator struct {
	Client  *clientv3.Client
	Session *concurrency.Session // 核心：基于 TTL 管理会话
	NodeID  string
}

func NewCoordinator(nodeID string, endpoints []string) (*Coordinator, error) {
	// 1. 初始化 Etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// 2. 创建 Session (它会自动维持心跳)
	// 一旦本程序崩溃，Session 过期，锁也会自动解开
	sess, err := concurrency.NewSession(cli, concurrency.WithTTL(TTL))
	if err != nil {
		return nil, err
	}

	return &Coordinator{
		Client:  cli,
		Session: sess,
		NodeID:  nodeID,
	}, nil
}

// Campaign 竞选 Master
// 这是一个阻塞方法，直到竞选成功才会返回 true
// 如果上下文取消，返回 false
func (c *Coordinator) Campaign(ctx context.Context) bool {
	// 使用 Etcd 官方的 Election 库
	e := concurrency.NewElection(c.Session, ElectionKey)

	log.Printf("[%s] Start campaigning for leader...", c.NodeID)

	// 阻塞调用：除非成功或者 ctx cancel，否则一直卡在这里等锁释放
	if err := e.Campaign(ctx, c.NodeID); err != nil {
		log.Printf("[%s] Campaign failed: %v", c.NodeID, err)
		return false
	}

	log.Printf("👑 [%s] is now the LEADER!", c.NodeID)
	return true
}

// RegisterNode 服务发现：告诉世界"我还活着"
func (c *Coordinator) RegisterNode() error {
	kv := clientv3.NewKV(c.Client)

	key := RegistryKey + c.NodeID
	value := "online" // 这里实际可以存放 IP:PORT

	// Put 时带上 Session 的 Lease
	// 只要 Session 在（程序没挂，网络没断），这个 key 就一直在
	// Session 断了，key 自动消失 -> 这就是自动摘除
	_, err := kv.Put(context.TODO(), key, value, clientv3.WithLease(c.Session.Lease()))

	log.Printf("[%s] Registered node to Etcd", c.NodeID)
	return err
}

func (c *Coordinator) Close() {
	c.Session.Close()
	c.Client.Close()
}
