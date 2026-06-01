package ctxutils

import (
	"context"

	"github.com/76Parker/golib/loglib"
)

const requestIdCtxKey = "request_id"
const loggerCtxKey = "logger"

// RequestID returns requestID from context
func RequestID(ctx context.Context) string {
	return ctx.Value(requestIdCtxKey).(string)
}

// SetRequestID set requestID in context and return new context with requestID
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIdCtxKey, requestID)
}

// Logger returns logger fron context
func Logger(ctx context.Context) loglib.Logger {
	return ctx.Value(loggerCtxKey).(loglib.Logger)
}

// SetLogger set loglib.Logger into context and returns new context with logger
func SetLogger(ctx context.Context, logger loglib.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}
