package handler

import "context"

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(UserIDKey).(int)
	return id, ok
}
