package handler

import "context"

type contextKey string

const authUserIDKey contextKey = "auth_user_id"
const authAdminIDKey contextKey = "auth_admin_id"

func withAuthUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, authUserIDKey, userID)
}

func currentUserID(ctx context.Context) int64 {
	value := ctx.Value(authUserIDKey)
	userID, _ := value.(int64)
	return userID
}

func withAuthAdminID(ctx context.Context, adminID int64) context.Context {
	return context.WithValue(ctx, authAdminIDKey, adminID)
}

func currentAdminID(ctx context.Context) int64 {
	value := ctx.Value(authAdminIDKey)
	adminID, _ := value.(int64)
	return adminID
}
