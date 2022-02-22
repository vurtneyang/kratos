package http_client

import "context"

type userKey struct{}

// NewContext adds Claims as a part of given context and returns a new one.
func NewContext(ctx context.Context, userId int64) context.Context {
	return context.WithValue(ctx, userKey{}, userId)
}

// FromContext returns Claims from current context, nil if not exists.
func FromContext(ctx context.Context) (userId int64, ok bool) {
	userId, ok = ctx.Value(userKey{}).(int64)
	return
}
