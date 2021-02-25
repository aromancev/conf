package trace

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
)

const (
	Key = "traceId"
)

func New(ctx context.Context, trace string) context.Context {
	return context.WithValue(ctx, Key, trace)
}

func Ctx(ctx context.Context) (context.Context, string) {
	var trace string
	if t, ok := ctx.Value(Key).(string); ok {
		trace = t
	} else {
		trace = uuid.New().String()
	}
	l := log.Ctx(ctx).With().Str(Key, trace).Logger()
	l.WithContext(ctx)
	return New(ctx, trace), trace
}

func Wrap(h func(http.ResponseWriter, *http.Request, httprouter.Params)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx, _ := Ctx(r.Context())
		r = r.WithContext(ctx)
		h(w, r, ps)
	}
}
