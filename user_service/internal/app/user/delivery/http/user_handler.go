package http

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"user_service/internal/app/user/delivery/dto"
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
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid user ID")
		utils.HandleError(ctx, utils.NewCustomError(http.StatusBadRequest, "Invalid user ID", err), h.log)
		return
	}
	user, err := h.userService.GetUserById(ctx.Request.Context(), id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "user_id": id, "path": ctx.Request.URL.Path}).Error("Failed to get user")
		utils.HandleError(ctx, err, h.log)
		return
	}
	resp := &dto.UserViewResponse{
		Id:    id,
		Name:  user.Name,
		Email: user.Email,
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetFilteredUserList(ctx *gin.Context) {

}

func (h *UserHandler) PostUser(ctx *gin.Context) {

}

func (h *UserHandler) PutUser(ctx *gin.Context) {

}

func (h *UserHandler) DeleteUser(ctx *gin.Context) {

}
