package middlewares

import (
	"context"
	"net/http"
	"sekai/internal/api/apierrs"
	"sekai/internal/entities/auth"
	"sekai/internal/entities/dto"
	"sekai/pkg/ctxutils"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserRepository interface {
	GetOrCreateUserFromKeycloak(ctx context.Context, user auth.KeycloakUser) (auth.User, error)
}

const authHeader = "Authorization"

func NoAuth(authRepo UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log := ctxutils.Logger(ctx.Request.Context())

		u, err := authRepo.GetOrCreateUserFromKeycloak(ctx.Request.Context(), auth.KeycloakUser{
			KeycloakSub: "local-dev-user",
			Email:       "local-dev@sekai.local",
			Username:    "local-dev",
		})
		if err != nil {
			log.Error("failed to get or create local user", "error", err)
			ctx.Error(apierrs.ErrInternalServerError())
			ctx.Abort()
			return
		}

		ctx.Request = ctx.Request.WithContext(ctxutils.SetUser(ctx.Request.Context(), u))
		ctx.Next()
	}
}

func Auth(jwks keyfunc.Keyfunc, issuer string, authRepo UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log := ctxutils.Logger(ctx.Request.Context())
		header := ctx.GetHeader(authHeader)
		if header == "" {
			log.Warn("no `Authorization` header", "code", http.StatusUnauthorized)
			ctx.Error(apierrs.ErrUnauthorized("authorization header is required"))
			ctx.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			log.Warn("authorization header not Bearer", "code", http.StatusUnauthorized)
			ctx.Error(apierrs.ErrUnauthorized("authorization header must be Bearer {token}"))
			ctx.Abort()
			return
		}

		tokenInBase64 := strings.TrimSpace(parts[1])

		claims := new(dto.KeycloakClaims)

		token, err := jwt.ParseWithClaims(
			tokenInBase64,
			claims,
			jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithIssuer(issuer),
		)
		if err != nil || !token.Valid {
			log.Warn("invalid token", "code", http.StatusUnauthorized, "error", err.Error())
			ctx.Error(apierrs.ErrUnauthorized("invalid token"))
			ctx.Abort()
			return
		}
		// MAY BE OPTIMIZE REQUESTS
		u, err := authRepo.GetOrCreateUserFromKeycloak(ctx.Request.Context(), auth.KeycloakUser{
			KeycloakSub: claims.Subject,
			Email:       claims.Email,
			Username:    claims.PreferredUsername,
		})

		if err != nil {
			log.Warn("failed to get or create user", "code", http.StatusUnauthorized, "error", err.Error())
			ctx.Error(apierrs.ErrUnauthorized("failed to get or create user"))
			ctx.Abort()
			return
		}
		ctx.Request = ctx.Request.WithContext(ctxutils.SetUser(ctx.Request.Context(), u))
		ctx.Next()
	}
}
