// Package authctx carries the authenticated user through request contexts
package authctx

import "context"

type contextKey struct{}

// WithUser returns a context carrying the authenticated username
func WithUser(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, contextKey{}, username)
}

// Username returns the authenticated username, or "" if absent
func Username(ctx context.Context) string {
	if v, ok := ctx.Value(contextKey{}).(string); ok {
		return v
	}
	return ""
}
