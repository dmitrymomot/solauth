package solauth

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
)

type (
	verifier interface {
		VerifyToken(tokenString string) (*Claims, error)
	}

	contextKey struct{ name string }
)

// TokenClaimsContextKey is the key for the token claims in the request context.
var TokenClaimsContextKey = &contextKey{name: "token-claims"}

// Get claims from request context
func GetClaimsFromRequest(r *http.Request) *Claims {
	claims, ok := r.Context().Value(TokenClaimsContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// Get claims from context
func GetClaimsFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(TokenClaimsContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// Middleware is a middleware for SolAuth.
// It will check the request for a valid token and
// add the claims to the request context.
func Middleware(v verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from request
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := v.VerifyToken(token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), TokenClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GoKitMiddleware is a middleware for SolAuth.
// It will check the context for a valid token and
// add the claims to the context.
func GoKitMiddleware(v verifier) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			token, ok := ctx.Value(jwt.JWTContextKey).(string)
			if !ok {
				return nil, ErrUnauthorized
			}

			// Validate token
			claims, err := v.VerifyToken(token)
			if err != nil {
				return nil, ErrUnauthorized
			}

			// Add user to request context
			ctx = context.WithValue(ctx, TokenClaimsContextKey, claims)
			return next(ctx, request)
		}
	}
}
