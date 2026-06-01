package ctxutils

import (
	"context"
	"sekai/internal/entities/auth"
)

const userInfoKey = "user_info"

func User(ctx context.Context) (auth.User, bool) {
	userInfo, ok := ctx.Value(userInfoKey).(auth.User)
	return userInfo, ok
}

func SetUser(ctx context.Context, userInfo auth.User) context.Context {
	return context.WithValue(ctx, userInfoKey, userInfo)
}
