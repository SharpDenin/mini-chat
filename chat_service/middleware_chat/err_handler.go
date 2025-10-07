package middleware_chat

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ErrorResponse представляет ошибку, которая возвращается клиенту
type ErrorResponse struct {
	Error string `json:"error" binding:"omitempty" example:"Invalid request parameters"`
}

func ErrorHandler(message string) *ErrorResponse {
	return &ErrorResponse{
		Error: message,
	}
}

type CustomError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func NewCustomError(statusCode int, message string, err error) *CustomError {
	return &CustomError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

func HandleError(ctx *gin.Context, err error, logger *logrus.Logger) {
	ctx.Header("Content-Type", "application/json")
	statusCode := http.StatusInternalServerError
	message := "Internal Server Error"

	switch e := err.(type) {
	case *CustomError:
		statusCode = e.StatusCode
		message = e.Message
		if e.Err != nil {
			logger.WithFields(logrus.Fields{"error": e.Err, "status": statusCode}).Error(message)
		} else {
			logger.WithFields(logrus.Fields{"status": statusCode}).Error(message)
		}
	case error:
		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusNotFound
			message = "Resource not found"
			logger.WithFields(logrus.Fields{"error": err, "status": statusCode}).Info(message)
		} else {
			logger.WithFields(logrus.Fields{"error": err, "status": statusCode}).Error(message)
		}
	}
	response := ErrorHandler(message)
	ctx.JSON(statusCode, response)
}

func ErrorMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) > 0 {
			for _, err := range ctx.Errors {
				HandleError(ctx, err, logger)
				return
			}
		}
	}
}
