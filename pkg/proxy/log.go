package proxy

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type reqLogKey struct{}

func withLogger(ctx context.Context, logger *log.Entry) context.Context {
	return context.WithValue(ctx, reqLogKey{}, logger)
}

func loggerFromContext(ctx context.Context) (*log.Entry, bool) {
	logger, ok := ctx.Value(reqLogKey{}).(*log.Entry)
	return logger, ok
}

func ensureLoggerFromContext(ctx context.Context) *log.Entry {
	logger, ok := loggerFromContext(ctx)
	if !ok {
		logger = log.NewEntry(log.StandardLogger())
	}
	return logger
}
