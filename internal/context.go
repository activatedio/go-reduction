package internal

import "context"

type contextKey struct {
	slug string
}

var (
	sessionIDKey = &contextKey{"session-id"}
)

func MustGetSessionID(ctx context.Context) string {
	return ctx.Value(sessionIDKey).(string)
}

func WithSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sid)
}
