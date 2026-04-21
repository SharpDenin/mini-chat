package http

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"profile_service/http/api_dto"
	"profile_service/internal/relation/service/helpers"
	"profile_service/internal/relation/service/interfaces"
	userService "profile_service/internal/user/service"
	"profile_service/middleware_profile"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type FriendshipHandler struct {
	userService       userService.UserServiceInterface
	friendshipService interfaces.FriendshipServiceInterface
	relationChecker   interfaces.UserRelationCheckerInterface
	log               *logrus.Logger
}

func NewFriendshipHandler(userService userService.UserServiceInterface, friendshipService interfaces.FriendshipServiceInterface,
	relationChecker interfaces.UserRelationCheckerInterface, log *logrus.Logger) *FriendshipHandler {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &FriendshipHandler{
		userService:       userService,
		friendshipService: friendshipService,
		relationChecker:   relationChecker,
		log:               log,
	}
}

// Friend Request

// PostFriendRequest
// @Summary Отправить запрос в друзья
// @Description Отправляет запрос на добавление в друзья другому пользователю
// @Tags Friends
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body api_dto.SendFriendRequestRequest true "Данные запроса"
// @Success 200 {object} api_dto.SuccessResponse "Запрос успешно отправлен"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Failure 403 {object} middleware_profile.ErrorResponse "Пользователь заблокирован"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден"
// @Failure 409 {object} middleware_profile.ErrorResponse "Конфликт (уже друзья или есть запрос)"
// @Router /friends/requests [post]
func (h *FriendshipHandler) PostFriendRequest(ctx *gin.Context) {
	var req api_dto.SendFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid friend request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	err := h.friendshipService.SendFriendRequest(ctx, req.ReceiverId, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, helpers.ErrUserNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User not found", err), h.log)
		case errors.Is(err, helpers.ErrAlreadyFriends), errors.Is(err, helpers.ErrFriendRequestExists):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusConflict, err.Error(), err), h.log)
		case errors.Is(err, helpers.ErrBlockedByUser):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusForbidden, err.Error(), err), h.log)
		case errors.Is(err, helpers.ErrCannotFriendYourself):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, err.Error(), err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.SuccessResponse{
		Success: true,
		Message: "Friend request sent successfully",
	})
}

// AnswerFriendRequest
// @Summary Ответить на запрос в друзья
// @Description Принимает или отклоняет запрос в друзья
// @Tags Friends
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request_id path int true "Id запроса"
// @Param request body api_dto.AnswerFriendRequestRequest true "Ответ на запрос"
// @Success 200 {object} api_dto.SuccessResponse "Ответ успешно обработан"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Failure 404 {object} middleware_profile.ErrorResponse "Запрос не найден"
// @Router /friends/requests/{request_id} [put]
func (h *FriendshipHandler) AnswerFriendRequest(ctx *gin.Context) {
	requestId, err := parseAndValidateID(ctx, "request_id", true)
	if err != nil {
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	var req api_dto.AnswerFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid answer request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	err = h.friendshipService.AnswerFriendRequest(ctx, requestId, req.Accept)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":      err,
			"request_id": requestId,
			"accept":     req.Accept,
			"path":       ctx.Request.URL.Path,
		}).Error("Failed to answer friend request")
		switch {
		case errors.Is(err, helpers.ErrFriendRequestNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "Friend request not found or already processed", err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	message := "Friend request rejected"
	if req.Accept {
		message = "Friend request accepted"
	}

	ctx.JSON(http.StatusOK, api_dto.SuccessResponse{
		Success: true,
		Message: message,
	})
}

// CancelFriendRequest
// @Summary Отменить запрос в друзья
// @Description Отменяет отправленный запрос в друзья
// @Tags Friends
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request_id path int true "Id запроса"
// @Success 200 {object} api_dto.SuccessResponse "Запрос успешно отменен"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Failure 404 {object} middleware_profile.ErrorResponse "Запрос не найден"
// @Router /friends/requests/{request_id} [delete]
func (h *FriendshipHandler) CancelFriendRequest(ctx *gin.Context) {
	requestId, err := parseAndValidateID(ctx, "request_id", true)
	if err != nil {
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	err = h.friendshipService.CancelFriendRequest(ctx, requestId)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":      err,
			"request_id": requestId,
			"path":       ctx.Request.URL.Path,
		}).Error("Failed to cancel friend request")
		switch {
		case errors.Is(err, helpers.ErrFriendRequestNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "Friend request not found or already processed", err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.SuccessResponse{
		Success: true,
		Message: "Friend request cancelled successfully",
	})
}

// CheckRequestState
// @Summary Проверить статус запроса в друзья
// @Description Проверяет статус запроса между текущим пользователем и указанным
// @Tags Friends
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param target_id query int true "Id целевого пользователя"
// @Success 200 {object} api_dto.RequestStateResponse "Статус запроса"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден"
// @Router /friends/requests/state [get]
func (h *FriendshipHandler) CheckRequestState(ctx *gin.Context) {
	targetId, err := parseAndValidateID(ctx, "target_id", false)
	if err != nil {
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	status, err := h.friendshipService.CheckRequestState(ctx, targetId)
	if err != nil {
		switch {
		case errors.Is(err, helpers.ErrUserNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User not found", err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.RequestStateResponse{
		Status: status,
	})
}

// GetFriendList
// @Summary Получить список друзей
// @Description Возвращает список друзей пользователя
// @Tags Friends
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} api_dto.FriendListResponse "Список друзей"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Router /friends [get]
func (h *FriendshipHandler) GetFriendList(ctx *gin.Context) {

	friends, err := h.friendshipService.GetFriendList(ctx)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error": err,
			"path":  ctx.Request.URL.Path,
		}).Error("Failed to get friend list")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusOK, friends)
}

// DeleteFriend
// @Summary Удалить из друзей
// @Description Удаляет пользователя из списка друзей
// @Tags Friends
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param friend_id path int true "Id друга"
// @Success 200 {object} api_dto.SuccessResponse "Друг успешно удален"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден или не являются друзьями"
// @Router /friends/{friend_id} [delete]
func (h *FriendshipHandler) DeleteFriend(ctx *gin.Context) {
	friendId, err := parseAndValidateID(ctx, "friend_id", true)
	if err != nil {
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	err = h.friendshipService.DeleteFromFriendList(ctx, friendId)
	if err != nil {
		switch {
		case errors.Is(err, helpers.ErrUserNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User not found", err), h.log)
		case errors.Is(err, helpers.ErrUsersNotFriends):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "Users are not friends", err), h.log)
		case errors.Is(err, helpers.ErrCannotDeleteYourself):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, err.Error(), err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.SuccessResponse{
		Success: true,
		Message: "Friend removed successfully",
	})
}

// BlockUser
// @Summary Заблокировать пользователя
// @Description Блокирует указанного пользователя
// @Tags Block
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body api_dto.BlockUserRequest true "Данные для блокировки"
// @Success 200 {object} api_dto.SuccessResponse "Пользователь заблокирован"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Failure 403 {object} middleware_profile.ErrorResponse "Попытка заблокировать себя"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден"
// @Failure 409 {object} middleware_profile.ErrorResponse "Пользователь уже заблокирован"
// @Router /block [post]
func (h *FriendshipHandler) BlockUser(ctx *gin.Context) {
	var req api_dto.BlockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid block request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	err := h.friendshipService.BlockUser(ctx, req.BlockedId, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, helpers.ErrUserNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User not found", err), h.log)
		case errors.Is(err, helpers.ErrUserAlreadyBlocked):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusConflict, "User already blocked", err), h.log)
		case errors.Is(err, helpers.ErrCannotBlockYourself):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusForbidden, err.Error(), err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.SuccessResponse{
		Success: true,
		Message: "User blocked successfully",
	})
}

// UnblockUser
// @Summary Разблокировать пользователя
// @Description Разблокирует ранее заблокированного пользователя
// @Tags Block
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body api_dto.UnblockUserRequest true "Данные для разблокировки"
// @Success 200 {object} api_dto.SuccessResponse "Пользователь разблокирован"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Failure 403 {object} middleware_profile.ErrorResponse "Попытка разблокировать себя"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден или блокировка отсутствует"
// @Router /block [delete]
func (h *FriendshipHandler) UnblockUser(ctx *gin.Context) {
	var req api_dto.UnblockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid unblock request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	err := h.friendshipService.UnblockUser(ctx, req.BlockedId)
	if err != nil {
		switch {
		case errors.Is(err, helpers.ErrUserNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User not found", err), h.log)
		case errors.Is(err, helpers.ErrUserNotBlocked):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User is not blocked", err), h.log)
		case errors.Is(err, helpers.ErrCannotUnblockYourself):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusForbidden, err.Error(), err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.SuccessResponse{
		Success: true,
		Message: "User unblocked successfully",
	})
}

// GetBlockInfo
// @Summary Получить информацию о блокировке
// @Description Проверяет, заблокирован ли пользователь
// @Tags Block
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param blocked_id path int true "Id проверяемого пользователя"
// @Success 200 {object} api_dto.BlockInfoResponse "Информация о блокировке"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Router /block/{blocked_id} [get]
func (h *FriendshipHandler) GetBlockInfo(ctx *gin.Context) {
	blockedId, err := parseAndValidateID(ctx, "blocked_id", true)
	if err != nil {
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	isBlocked, err := h.relationChecker.CheckUserIsBlocked(ctx, blockedId)
	if err != nil {
		switch {
		case errors.Is(err, helpers.ErrUserNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User not found", err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.BlockInfoResponse{
		IsBlocked: isBlocked,
		BlockedId: blockedId,
	})
}

// CheckAreFriends
// @Summary Проверить статус дружбы
// @Description Проверяет, являются ли пользователи друзьями
// @Tags Friends
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id1 query int true "Первый пользователь"
// @Param user_id2 query int true "Второй пользователь"
// @Success 200 {object} api_dto.AreFriendsResponse "Результат проверки"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Router /friends/check [get]
func (h *FriendshipHandler) CheckAreFriends(ctx *gin.Context) {
	userId1, err := parseAndValidateID(ctx, "user_id1", false)
	if err != nil {
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}
	userId2, err := parseAndValidateID(ctx, "user_id2", false)
	if err != nil {
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	areFriends, err := h.relationChecker.CheckUsersAreFriends(ctx, userId1, userId2)
	if err != nil {
		switch {
		case errors.Is(err, helpers.ErrUserNotFound):
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusNotFound, "User not found", err), h.log)
		default:
			middleware_profile.HandleError(ctx, err, h.log)
		}
		return
	}

	ctx.JSON(http.StatusOK, api_dto.AreFriendsResponse{
		AreFriends: areFriends,
	})
}

func parseAndValidateID(ctx *gin.Context, paramName string, isPath bool) (int64, error) {
	var idStr string
	if isPath {
		idStr = ctx.Param(paramName)
	} else {
		idStr = ctx.Query(paramName)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, middleware_profile.NewCustomError(http.StatusBadRequest,
			fmt.Sprintf("Invalid %s: must be positive integer", paramName), err)
	}
	return id, nil
}
