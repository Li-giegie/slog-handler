# slog-handler

`log/slog` Handler 实现集合。每个实现一个文件，独立使用，按需取用。

## 实现列表

| Handler | 文件 | 说明 |
|---------|------|------|
| MultiHandler | `multi_handler.go` | 多处理器扇出包装器，见下文 |

## MultiHandler

把一条日志**扇出（fan-out）到多个 Handler**，各自按自己的级别、格式、目的地输出。标准库没有内置多输出，典型场景：

- 文本格式进 stdout（开发调试，DEBUG 级起）
- JSON 格式进文件/采集器（生产采集，WARN 级起）
- 或再加自定义 Handler（采样、脱敏、染色…）

### 用法

```go
import slog_multihandler "server/.temp/slog-handler"

text := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
jsonH := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})

slog.SetDefault(slog.New(slog_multihandler.NewMultiHandler(text, jsonH)))

slog.Debug("只进文本 handler")           // JSON 的 WARN 起步，Enabled 过滤掉
slog.Warn("两个 handler 都输出")          // 文本 + JSON 双输出

logger := slog.Default().With("service", "stwo")   // 属性转发到所有子 handler
logger.WithGroup("request").Warn("组内", slog.String("id", "123"))
```

### 语义

| 方法 | 行为 |
|------|------|
| `Enabled` | 任一子 handler 启用即启用 |
| `Handle` | 逐个分发给已启用的子 handler；用 `errors.Join` **聚合所有错误**，不因单个失败中断其余 handler |
| `WithAttrs` | 转发给每个子 handler，返回**新的** MultiHandler；`len(attrs)==0` 时返回自身 |
| `WithGroup` | 转发给每个子 handler，返回**新的** MultiHandler；`name==""` 时返回自身 |

### 实现要点

1. **无内部状态**：不缓存 attrs/group，`With*` 直接转发并派生新实例——slog 约定 `With*` 的返回值是不可变的派生 Handler，**不能原地修改**自身（否则已存在的派生 logger 行为会被意外改变）。
2. **错误聚合**：一个子 Handler 写失败（如磁盘满）不影响其余输出，最终错误合并返回。
3. **不 `Clone` Record**：slog 契约是"会修改 Record 的实现必须先 Clone"；包装器不修改，直接转发安全且省一次分配。
4. **并发安全**：`Handle` 串行遍历子 handler，依赖各子 Handler 自身的并发安全保证（标准 Handler 均满足）。
5. 空组名/空属性的 `With*` 返回自身，避免无意义派生（slog 文档明确约定）。

### 测试

```shell
go test  -v
```

## 新增 Handler 约定

- 一个实现一个文件：`xxx_handler.go` + `xxx_handler_test.go`
- 遵守 slog.Handler 契约：`Handle` 并发安全、`With*` 返回不可变派生、会改 Record 就先 `Clone`
- 在上方实现列表中登记一行
