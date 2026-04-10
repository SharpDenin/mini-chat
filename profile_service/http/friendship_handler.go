package http

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"profile_service/http/api_dto"
	"profile_service/internal/relation/service/interfaces"
	userService "profile_service/internal/user/service"
	"profile_service/middleware_profile"
	"strconv"
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
// @Failure 409 {object} middleware_profile.ErrorResponse "Конфликт (уже друзья или есть запрос)"
// @Router /friends/requests [post]
func (h *FriendshipHandler) PostFriendRequest(ctx *gin.Context) {
	var req api_dto.SendFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid friend request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	senderId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	senderIdInt64, ok := senderId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	err := h.friendshipService.SendFriendRequest(ctx.Request.Context(), senderIdInt64, req.ReceiverId, req.Message)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":       err,
			"sender_id":   senderIdInt64,
			"receiver_id": req.ReceiverId,
			"path":        ctx.Request.URL.Path,
		}).Error("Failed to send friend request")
		middleware_profile.HandleError(ctx, err, h.log)
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
	requestId, err := strconv.ParseInt(ctx.Param("request_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request ID", err), h.log)
		return
	}

	var req api_dto.AnswerFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid answer request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	userId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	userIdInt64, ok := userId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	err = h.friendshipService.AnswerFriendRequest(ctx.Request.Context(), requestId, userIdInt64, req.Accept)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":      err,
			"request_id": requestId,
			"user_id":    userIdInt64,
			"accept":     req.Accept,
			"path":       ctx.Request.URL.Path,
		}).Error("Failed to answer friend request")
		middleware_profile.HandleError(ctx, err, h.log)
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
	requestId, err := strconv.ParseInt(ctx.Param("request_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request ID", err), h.log)
		return
	}

	userId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	userIdInt64, ok := userId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	// Отмена запроса = отклонение (accept = false)
	err = h.friendshipService.AnswerFriendRequest(ctx.Request.Context(), requestId, userIdInt64, false)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":      err,
			"request_id": requestId,
			"user_id":    userIdInt64,
			"path":       ctx.Request.URL.Path,
		}).Error("Failed to cancel friend request")
		middleware_profile.HandleError(ctx, err, h.log)
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
// @Router /friends/requests/state [get]
func (h *FriendshipHandler) CheckRequestState(ctx *gin.Context) {
	targetId, err := strconv.ParseInt(ctx.Query("target_id"), 10, 64)
	if err != nil || targetId <= 0 {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid target ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid target_id", err), h.log)
		return
	}

	userId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	userIdInt64, ok := userId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	status, err := h.friendshipService.CheckRequestState(ctx.Request.Context(), userIdInt64, targetId)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":     err,
			"user_id":   userIdInt64,
			"target_id": targetId,
			"path":      ctx.Request.URL.Path,
		}).Error("Failed to check request state")
		middleware_profile.HandleError(ctx, err, h.log)
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
// @Param user_id query int false "Id пользователя (если не указан, возвращает для текущего)"
// @Success 200 {object} api_dto.FriendListResponse "Список друзей"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Не авторизован"
// @Router /friends [get]
func (h *FriendshipHandler) GetFriendList(ctx *gin.Context) {
	var userIdInt64 int64
	userIdParam := ctx.Query("user_id")

	if userIdParam != "" {
		id, err := strconv.ParseInt(userIdParam, 10, 64)
		if err != nil {
			h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid user_id parameter")
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid user_id", err), h.log)
			return
		}
		userIdInt64 = id
	} else {
		// Иначе берем текущего пользователя
		userId, exists := ctx.Get("user_id")
		if !exists {
			h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
			return
		}
		var ok bool
		userIdInt64, ok = userId.(int64)
		if !ok {
			h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
			middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
			return
		}
	}

	friends, err := h.friendshipService.GetFriendList(ctx.Request.Context(), userIdInt64)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":   err,
			"user_id": userIdInt64,
			"path":    ctx.Request.URL.Path,
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
// @Failure 404 {object} middleware_profile.ErrorResponse "Друг не найден"
// @Router /friends/{friend_id} [delete]
func (h *FriendshipHandler) DeleteFriend(ctx *gin.Context) {
	friendId, err := strconv.ParseInt(ctx.Param("friend_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid friend ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid friend ID", err), h.log)
		return
	}

	userId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	userIDInt64, ok := userId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	err = h.friendshipService.DeleteFromFriendList(ctx.Request.Context(), userIDInt64, friendId)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":     err,
			"user_id":   userIDInt64,
			"friend_id": friendId,
			"path":      ctx.Request.URL.Path,
		}).Error("Failed to delete friend")
		middleware_profile.HandleError(ctx, err, h.log)
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
// @Failure 409 {object} middleware_profile.ErrorResponse "Пользователь уже заблокирован"
// @Router /block [post]
func (h *FriendshipHandler) BlockUser(ctx *gin.Context) {
	var req api_dto.BlockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid block request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	// Получаем ID блокирующего из контекста
	blockerId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	blockerIdInt64, ok := blockerId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	err := h.friendshipService.BlockUser(ctx.Request.Context(), blockerIdInt64, req.BlockedId, req.Reason)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":      err,
			"blocker_id": blockerIdInt64,
			"blocked_id": req.BlockedId,
			"path":       ctx.Request.URL.Path,
		}).Error("Failed to block user")
		middleware_profile.HandleError(ctx, err, h.log)
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
// @Failure 404 {object} middleware_profile.ErrorResponse "Блокировка не найдена"
// @Router /block [delete]
func (h *FriendshipHandler) UnblockUser(ctx *gin.Context) {
	var req api_dto.UnblockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid unblock request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	// Получаем ID блокирующего из контекста
	blockerId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	blockerIdInt64, ok := blockerId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	err := h.friendshipService.UnblockUser(ctx.Request.Context(), blockerIdInt64, req.BlockedId)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":      err,
			"blocker_id": blockerIdInt64,
			"blocked_id": req.BlockedId,
			"path":       ctx.Request.URL.Path,
		}).Error("Failed to unblock user")
		middleware_profile.HandleError(ctx, err, h.log)
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
	blockedId, err := strconv.ParseInt(ctx.Param("blocked_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid blocked ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid blocked_id", err), h.log)
		return
	}

	// Получаем ID блокирующего из контекста
	blockerId, exists := ctx.Get("user_id")
	if !exists {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("User ID not found in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "User not authenticated", nil), h.log)
		return
	}

	blockerIdInt64, ok := blockerId.(int64)
	if !ok {
		h.log.WithField("path", ctx.Request.URL.Path).Warn("Invalid user ID type in context")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusInternalServerError, "Invalid user ID format", nil), h.log)
		return
	}

	isBlocked, err := h.relationChecker.CheckUserIsBlocked(ctx.Request.Context(), blockerIdInt64, blockedId)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":      err,
			"blocker_id": blockerIdInt64,
			"blocked_id": blockedId,
			"path":       ctx.Request.URL.Path,
		}).Error("Failed to check block info")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusOK, api_dto.BlockInfoResponse{
		IsBlocked: isBlocked,
		BlockerId: blockerIdInt64,
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
	userId1, err := strconv.ParseInt(ctx.Query("user_id1"), 10, 64)
	if err != nil || userId1 <= 0 {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid user_id1")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid user_id1", err), h.log)
		return
	}

	userId2, err := strconv.ParseInt(ctx.Query("user_id2"), 10, 64)
	if err != nil || userId2 <= 0 {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid user_id2")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid user_id2", err), h.log)
		return
	}

	areFriends, err := h.relationChecker.CheckUsersAreFriends(ctx.Request.Context(), userId1, userId2)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error":    err,
			"user_id1": userId1,
			"user_id2": userId2,
			"path":     ctx.Request.URL.Path,
		}).Error("Failed to check friendship")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusOK, api_dto.AreFriendsResponse{
		AreFriends: areFriends,
	})
}
