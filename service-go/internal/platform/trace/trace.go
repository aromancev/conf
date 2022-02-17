package trace

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ctxKey struct{}

func New(ctx context.Context, trace string) context.Context {
	l := log.Logger.With().Str(logKey, trace).Logger()
	ctx = l.WithContext(ctx)
	return context.WithValue(ctx, ctxKey{}, trace)
}

func Ctx(ctx context.Context) (context.Context, string) {
	var trace string
	if t, ok := ctx.Value(ctxKey{}).(string); ok {
		trace = t
	} else {
		trace = uuid.New().String()
	}
	l := log.Logger.With().Str(logKey, trace).Logger()
	ctx = l.WithContext(ctx)
	return New(ctx, trace), trace
}

func ID(ctx context.Context) string {
	_, traceID := Ctx(ctx)
	return traceID
}

const (
	logKey = "traceId"
)
