package utils

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
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

	switch err := err.(type) {
	case *CustomError:
		statusCode = err.StatusCode
		message = err.Message
		logger.Errorf("%s: %v", message, err.Err)
	case error:
		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusNotFound
			message = "Resource not found"
			logger.Infof("Resource not found: %v", err)
		} else {
			logger.Errorf("Unexpected error: %v", err)
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
