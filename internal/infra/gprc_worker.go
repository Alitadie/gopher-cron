package infra

import (
	"context"
	"gopher-cron/api/proto" // 引入生成的包
	"gopher-cron/internal/core"
	"log"
	"net"

	"google.golang.org/grpc"
)

// GrpcServer 适配层：将 gRPC 流量转换为核心域调用
type GrpcWorker struct {
	proto.UnimplementedExecutorServer // 必须嵌入这个 forward compatible 结构
	jobManager                        *core.JobManager
	port                              string
}

func NewGrpcWorker(port string, mgr *core.JobManager) *GrpcWorker {
	return &GrpcWorker{
		port:       port,
		jobManager: mgr,
	}
}

// PushJob 实现 proto 中定义的接口
func (s *GrpcWorker) PushJob(ctx context.Context, req *proto.PushJobRequest) (*proto.PushJobResponse, error) {
	log.Printf("[gRPC] Received Job: %s (ID: %s)", req.JobName, req.JobId)

	// DDD: 应用层调用领域层
	// 这里我们需要稍微改造 core.JobManager 让它支持动态执行一次性任务
	// 或者我们将收到 Params 作为参数传入已注册的任务

	// 这里演示简单模式：根据 ID 触发已注册任务，实际项目应该用 req.JobParam 动态加载上下文
	s.jobManager.TriggerJob(req.JobId)

	return &proto.PushJobResponse{
		Code: 200,
		Msg:  "Success received",
	}, nil
}

// Start 启动 TCP 监听
func (s *GrpcWorker) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterExecutorServer(grpcServer, s)

	log.Printf("Worker gRPC server listening on :%s", s.port)
	return grpcServer.Serve(lis)
}
