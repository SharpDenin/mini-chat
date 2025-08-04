package http

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"user_service/internal/app/user/service"
	"user_service/internal/utils"
)

type UserHandler struct {
	userService service.UserServiceInterface
	log         *logrus.Logger
}

func NewUserHandler(userService service.UserServiceInterface, log *logrus.Logger) *UserHandler {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

func (h *UserHandler) GetUserById(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 0, 64)
	if err != nil {
		h.log.WithError(err).Debug("Invalid ID parameter")
		ctx.JSON(http.StatusBadRequest, utils.ErrorHandler("Invalid ID parameter"))
		return
	}
	user, err := h.userService.GetUserById(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.log.WithField("id", id).Debug("User not found")
			ctx.JSON(http.StatusNotFound, utils.ErrorHandler("User not found"))
			return
		}
		h.log.WithError(err).Debug("User not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.log.WithField("id", id).Info("User retrieved successfully")
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) GetFilteredUserList(ctx *gin.Context) {

}

func (h *UserHandler) PostUser(ctx *gin.Context) {

}

func (h *UserHandler) PutUser(ctx *gin.Context) {

}

func (h *UserHandler) DeleteUser(ctx *gin.Context) {

}
