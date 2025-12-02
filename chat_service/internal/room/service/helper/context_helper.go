package helper

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserIdFromContext(ctx context.Context) (int64, error) {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return 0, errors.New("context is not a gin context")
	}

	userId, exists := ginCtx.Get("user_id")
	if !exists {
		return 0, errors.New("user_id not found in context")
	}

	userIdStr, ok := userId.(string)
	if !ok {
		return 0, errors.New("user_id is not a string")
	}

	return strconv.ParseInt(userIdStr, 10, 64)
}
