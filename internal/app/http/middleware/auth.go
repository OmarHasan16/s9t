// internal/app/http/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"s9t.os/internal/platform/types"
)

// Context keys for propagating identity downstream
type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	ActorIDKey  contextKey = "actor_id"
	RoleIDKey   contextKey = "role_id"
)

// AuthMiddleware validates JWT and injects identity into context
func AuthMiddleware(secretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims := jwt.MapClaims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			tenantIDStr, _ := claims["tenant_id"].(string)
			actorIDStr, _ := claims["sub"].(string) // Standard JWT subject claim for User ID
			roleIDStr, _ := claims["role_id"].(string)

			if tenantIDStr == "" || actorIDStr == "" {
				http.Error(w, "invalid token claims: missing tenant or user identity", http.StatusUnauthorized)
				return
			}

			// Inject into request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, TenantIDKey, types.TenantID(tenantIDStr))
			ctx = context.WithValue(ctx, ActorIDKey, types.ActorID(actorIDStr))
			
			if roleIDStr != "" {
				ctx = context.WithValue(ctx, RoleIDKey, types.RoleID(roleIDStr))
			}

			// Pass to the next handler with the enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
