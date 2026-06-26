package logger

import (
	"context"
	"log/slog"
	"os"
)

var Log *slog.Logger

const CorrelationIDKey = "correlation_id"

func InitLogger() {
	// set default structured JSON handler to stdout
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// helper for log with context that automatically includes correalation ID if present in the context
func getLogArgs(ctx context.Context, args []any) []any {
	if ctx != nil {
		if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
			return append(args, slog.String("correlation_id", correlationID))
		}
	}
	return args
}

func Info(ctx context.Context, msg string, args ...any) {
	Log.InfoContext(ctx, msg, getLogArgs(ctx, args)...)
}

func Error(ctx context.Context, msg string, args ...any) {
	Log.ErrorContext(ctx, msg, getLogArgs(ctx, args)...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	Log.WarnContext(ctx, msg, getLogArgs(ctx, args)...)
}
