package middleware

import (
	"context"
	"s9t.os/internal/platform/types"
)

type tenantIDKey struct{}
type actorIDKey struct{}

// WithTenantID safely injects the TenantID into the context
func WithTenantID(ctx context.Context, tenantID types.TenantID) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

// TenantIDFromContext safely extracts the TenantID
func TenantIDFromContext(ctx context.Context) (types.TenantID, bool) {
	tenantID, ok := ctx.Value(tenantIDKey{}).(types.TenantID)
	return tenantID, ok
}

// WithActorID safely injects the ActorID into the context
func WithActorID(ctx context.Context, actorID types.ActorID) context.Context {
	return context.WithValue(ctx, actorIDKey{}, actorID)
}

// ActorIDFromContext safely extracts the ActorID
func ActorIDFromContext(ctx context.Context) (types.ActorID, bool) {
	actorID, ok := ctx.Value(actorIDKey{}).(types.ActorID)
	return actorID, ok
}
