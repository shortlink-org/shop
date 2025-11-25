package loggeradapter

import (
	"context"
	"log/slog"
	"os"

	sdklogger "github.com/shortlink-org/go-sdk/logger"
	shortlogger "github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/logger/field"
)

// Adapter implements shortlink's logger interface on top of go-sdk/logger.
type Adapter struct {
	inner sdklogger.Logger
}

// New wraps a go-sdk logger and exposes the legacy shortlink logger interface.
func New(inner sdklogger.Logger) shortlogger.Logger {
	return &Adapter{inner: inner}
}

func (a *Adapter) Fatal(msg string, fields ...field.Fields) {
	a.inner.Error(msg, convertFields(fields...)...)
	os.Exit(1)
}

func (a *Adapter) FatalWithContext(ctx context.Context, msg string, fields ...field.Fields) {
	a.inner.ErrorWithContext(ctx, msg, convertFields(fields...)...)
	os.Exit(1)
}

func (a *Adapter) Error(msg string, fields ...field.Fields) {
	a.inner.Error(msg, convertFields(fields...)...)
}

func (a *Adapter) ErrorWithContext(ctx context.Context, msg string, fields ...field.Fields) {
	a.inner.ErrorWithContext(ctx, msg, convertFields(fields...)...)
}

func (a *Adapter) Warn(msg string, fields ...field.Fields) {
	a.inner.Warn(msg, convertFields(fields...)...)
}

func (a *Adapter) WarnWithContext(ctx context.Context, msg string, fields ...field.Fields) {
	a.inner.WarnWithContext(ctx, msg, convertFields(fields...)...)
}

func (a *Adapter) Info(msg string, fields ...field.Fields) {
	a.inner.Info(msg, convertFields(fields...)...)
}

func (a *Adapter) InfoWithContext(ctx context.Context, msg string, fields ...field.Fields) {
	a.inner.InfoWithContext(ctx, msg, convertFields(fields...)...)
}

func (a *Adapter) Debug(msg string, fields ...field.Fields) {
	a.inner.Debug(msg, convertFields(fields...)...)
}

func (a *Adapter) DebugWithContext(ctx context.Context, msg string, fields ...field.Fields) {
	a.inner.DebugWithContext(ctx, msg, convertFields(fields...)...)
}

func (a *Adapter) Get() any {
	return a.inner
}

func (a *Adapter) Close() error {
	return a.inner.Close()
}

func convertFields(fields ...field.Fields) []slog.Attr {
	if len(fields) == 0 {
		return nil
	}

	attrs := make([]slog.Attr, 0, len(fields))
	for _, set := range fields {
		for key, value := range set {
			attrs = append(attrs, slog.Any(key, value))
		}
	}

	return attrs
}
