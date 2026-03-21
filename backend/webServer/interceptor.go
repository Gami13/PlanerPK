package webServer

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
)

type contextKey string

const authTokenKey contextKey = "AuthorizationToken"

func AuthInterceptor() connect.UnaryInterceptorFunc {
	// Procedures that require authentication
	authRequired := map[string]bool{
		"/pkplaner.v1.PKPlanerService/SetChosenColumn": true,
		"/pkplaner.v1.PKPlanerService/SetUserActive":   true,
		"/pkplaner.v1.PKPlanerService/GetUserInfo":     true,
	}

	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if !authRequired[req.Spec().Procedure] {
				return next(ctx, req)
			}

			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authorization header"))
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization header format"))
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
			if token == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token is empty"))
			}
			// Add the token to the context
			ctx = context.WithValue(ctx, authTokenKey, token)

			return next(ctx, req)
		})
	})
}
