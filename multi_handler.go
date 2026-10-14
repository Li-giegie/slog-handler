package slog_handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// MultiHandler 多处理器包装器：把一条日志扇出（fan-out）到多个 slog.Handler
// 设计要点：
//   - 不持有 attrs/group 状态，WithAttrs/WithGroup 直接转发给各子 handler 并返回新的 MultiHandler
//     （slog 约定：With* 返回不可变的派生 handler，不能原地修改自身）
//   - Enabled：任一子 handler 启用即启用
//   - Handle：逐个分发，聚合所有错误（不因单个失败中断其余 handler）
//   - 不 Clone record：子 handler 按契约不修改传入的 Record，直接转发安全
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r); err != nil {
			errs = append(errs, fmt.Errorf("handler %T: %w", h, err))
		}
	}
	return errors.Join(errs...)
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return m // slog 约定：无属性时返回等价 handler（可返回自身）
	}
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return m // slog 约定：空组名返回等价 handler
	}
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}
