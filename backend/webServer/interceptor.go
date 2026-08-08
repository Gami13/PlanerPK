package webServer

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/webServer/handler"
)

func AuthInterceptor(query db.Querier) connect.UnaryInterceptorFunc {
	// Procedures that require authentication
	authRequired := map[string]bool{
		"/pkplaner.v1.PKPlanerService/SetChosenColumn": true,
		"/pkplaner.v1.PKPlanerService/SetUserActive":   true,
		"/pkplaner.v1.PKPlanerService/GetUserInfo":     true,
		"/pkplaner.v1.PKPlanerService/GetClasses":      true,
		"/pkplaner.v1.PKPlanerService/GetOverrides":    true,
		"/pkplaner.v1.PKPlanerService/SetOverride":     true,
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

			user, err := query.GetUserByAccessToken(ctx, token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization token"))
			}

			// Add the user to the context
			ctx = context.WithValue(ctx, handler.AuthUserKey, user)

			return next(ctx, req)
		})
	})
}

func LoggingInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			resp, err := next(ctx, req)
			duration := time.Since(start)
			if err != nil {
				code := connect.CodeOf(err)
				switch code {
				case connect.CodeInvalidArgument,
					connect.CodeNotFound,
					connect.CodeUnauthenticated,
					connect.CodePermissionDenied,
					connect.CodeAlreadyExists,
					connect.CodeFailedPrecondition,
					connect.CodeResourceExhausted,
					connect.CodeCanceled:
					slog.Warn("RPC Request Client Warning",
						slog.String("procedure", req.Spec().Procedure),
						slog.String("code", code.String()),
						slog.Duration("duration", duration),
						slog.Any("error", err),
					)
				default:
					slog.Error("RPC Request Server Error",
						slog.String("procedure", req.Spec().Procedure),
						slog.String("code", code.String()),
						slog.Duration("duration", duration),
						slog.Any("error", err),
					)
				}
			} else {
				slog.Info("RPC Request Completed",
					slog.String("procedure", req.Spec().Procedure),
					slog.Duration("duration", duration),
				)
			}
			return resp, err
		}
	}
}
