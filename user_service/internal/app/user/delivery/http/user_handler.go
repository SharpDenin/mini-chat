package http

import (
	"net/http"
	"os"
	"strconv"
	"user_service/internal/app/user/delivery/dto"
	"user_service/internal/app/user/delivery/dto/mappers"
	"user_service/internal/app/user/service"
	"user_service/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	resp := mappers.ConvertToServiceUser(user)
	ctx.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetFilteredUserList(ctx *gin.Context) {
	var req *dto.UserFilterRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid filter request")
		utils.HandleError(ctx, utils.NewCustomError(http.StatusBadRequest, "Invalid filter request", err), h.log)
		return
	}
	filter := mappers.ConvertToServiceFilter(req)
	list, err := h.userService.GetAllUsers(ctx.Request.Context(), *filter)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to get users")
		utils.HandleError(ctx, err, h.log)
		return
	}
	resp := mappers.ConvertToServiceList(list)
	ctx.JSON(http.StatusOK, resp)
}

func (h *UserHandler) PostUser(ctx *gin.Context) {
	var req *dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid create request")
		utils.HandleError(ctx, utils.NewCustomError(http.StatusBadRequest, "Invalid create request", err), h.log)
		return
	}
	mappedReq := mappers.ConvertToServiceCreate(req)
	id, err := h.userService.CreateUser(ctx.Request.Context(), mappedReq)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to create user")
		utils.HandleError(ctx, err, h.log)
		return
	}
	ctx.JSON(http.StatusCreated, id)
}

func (h *UserHandler) PutUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid user ID")
		utils.HandleError(ctx, utils.NewCustomError(http.StatusBadRequest, "Invalid user ID", err), h.log)
		return
	}
	var req *dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid update request")
		utils.HandleError(ctx, utils.NewCustomError(http.StatusBadRequest, "Invalid update request", err), h.log)
		return
	}
	mappedReq := mappers.ConvertToServiceUpdate(req)
	err = h.userService.UpdateUser(ctx.Request.Context(), id, mappedReq)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to update user")
		utils.HandleError(ctx, err, h.log)
		return
	}
	ctx.JSON(http.StatusNoContent, id)
}

func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Invalid user ID")
		utils.HandleError(ctx, utils.NewCustomError(http.StatusBadRequest, "Invalid user ID", err), h.log)
		return
	}
	err = h.userService.DeleteUser(ctx.Request.Context(), id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to delete user")
		utils.HandleError(ctx, err, h.log)
		return
	}
	ctx.JSON(http.StatusNoContent, id)
}
