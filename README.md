# Gopher Cron

Gopher Cron 是一个轻量级、高并发的 Go 语言定时任务调度器。

## 特性

- **高并发执行**：支持大量任务并发执行。
- **轻量级**：核心代码简洁，易于集成。
- **Docker 支持**：提供 Dockerfile，方便容器化部署。

## 快速开始

### 运行测试

```bash
make test
```

### 运行项目

```bash
make run
```

### Docker 构建

```bash
docker build -t gopher-cron .
```

## 许可证

MIT
