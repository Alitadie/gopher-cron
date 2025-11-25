package infra

import (
	"context"
	"fmt"
	"gopher-cron/api/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MasterDispatcher 负责分发任务
type MasterDispatcher struct {
	// 实际场景应使用 Connection Pool 避免重复 TCP 握手
}

func NewMasterDispatcher() *MasterDispatcher {
	return &MasterDispatcher{}
}

// Dispatch 发送任务到指定 Worker
func (d *MasterDispatcher) Dispatch(ctx context.Context, workerIP string, jobID string, jobName string) error {
	// 1. 建立连接
	// WithTransportCredentials(insecure...) 用于内网非 TLS 环境
	conn, err := grpc.NewClient(workerIP, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close() // 注意：短连接模式消耗大，高并发下需维护长连接池

	// 2. 创建 Client Stub
	client := proto.NewExecutorClient(conn)

	// 3. 发送 RPC 请求
	// 设置独立的超时，防止 Worker 卡死拖垮 Master
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := client.PushJob(callCtx, &proto.PushJobRequest{
		JobId:   jobID,
		JobName: jobName,
		LogId:   time.Now().UnixNano(),
	})

	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf("worker return error: %s", resp.Msg)
	}

	return nil
}
