package slog_handler

import (
	"errors"
	"log/slog"
	"os"
	"testing"
)

func TestMultiHandler(t *testing.T) {
	// 演示：文本格式到 stdout（DEBUG 级起），JSON 格式到 stderr（WARN 级起）
	text := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	jsonH := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
	slog.SetDefault(slog.New(NewMultiHandler(text, jsonH)))

	slog.Debug("只进文本 handler")
	slog.Info("只进文本 handler")
	slog.Warn("两个 handler 都输出")
	slog.Error("两个 handler 都输出", slog.Any("err", errors.New("demo error")))

	// With 派生 logger：属性与组通过 MultiHandler 正确转发到两个子 handler
	logger := slog.Default().With(slog.String("service", "stwo"))
	logger.Warn("带公共属性")
	logger.WithGroup("request").Warn("组内日志", slog.String("id", "123"))

}
