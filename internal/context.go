package internal

import "context"

type contextKey struct {
	slug string
}

var (
	sessionIDKey = &contextKey{"session-id"}
)

func MustGetSessionID(ctx context.Context) string {
	raw := ctx.Value(sessionIDKey)
	if raw == nil {
		panic("no session id in context")
	}
	return raw.(string)
}

// SessionID answers the session key in ctx, if there is a non-empty one.
func SessionID(ctx context.Context) (string, bool) {
	sid, ok := ctx.Value(sessionIDKey).(string)
	return sid, ok && sid != ""
}

func WithSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sid)
}
