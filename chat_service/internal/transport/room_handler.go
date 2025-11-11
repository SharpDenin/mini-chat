package transport

import (
	"chat_service/internal/room/service"
	"chat_service/internal/transport/api_dto"
	"chat_service/internal/transport/room_mapper"
	"chat_service/middleware_chat"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RoomHandler struct {
	log               *logrus.Logger
	roomService       service.RoomServiceInterface
	roomMemberService service.RoomMemberServiceInterface
}

func NewRoomHandler(log *logrus.Logger, roomService service.RoomServiceInterface, roomMemberService service.RoomMemberServiceInterface) *RoomHandler {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &RoomHandler{
		log:               log,
		roomService:       roomService,
		roomMemberService: roomMemberService,
	}
}

// CreateRoom
// @Summary Создание комнаты
// @Description Создает комнату и возвращает ее Id
// @Tags Room
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body api_dto.CreateRoomRequest true "Данные для создания комнаты"
// @Success 201 {integer} int "Id созданной комнаты"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_chat.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room [post]
func (h *RoomHandler) CreateRoom(ctx *gin.Context) {
	var req *api_dto.CreateRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	id, err := h.roomService.CreateRoom(ctx, req.Name)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error creating room")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusCreated, id)
}

// GetRoom
// @Summary Получение комнаты
// @Description Возвращает информацию о комнате по ее Id
// @Tags Room
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Id комнаты"
// @Success 200 {object} api_dto.GetRoomResponse "Информация о комнате"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные данные"
// @Failure 404 {object} middleware_chat.ErrorResponse "Комната не найдена"
// @Failure 401 {object} middleware_chat.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room/{id} [get]
func (h *RoomHandler) GetRoom(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	room, err := h.roomService.GetRoomById(ctx, id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error getting room")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}
	resp := room_mapper.GetRoomToHandlerDto(room)

	ctx.JSON(http.StatusOK, resp)

}

// GetRoomList
// @Summary Получить список комнат
// @Description Возвращает список комнат
// @Tags Room
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param searchQuery query string false "Фильтр по имени"
// @Param limit query int true "Лимит (1-100)" minimum(1) maximum(100)
// @Param offset query int false "Смещение" minimum(0)
// @Success 200 {object} api_dto.GetRoomListResponse "Список комнат"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные параметры фильтрации"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room [get]
func (h *RoomHandler) GetRoomList(ctx *gin.Context) {
	var searchFilter *api_dto.SearchRoomRequest
	if err := ctx.ShouldBindQuery(&searchFilter); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	req := room_mapper.SearchQueryToServiceFilter(searchFilter)
	roomList, err := h.roomService.GetRoomList(ctx, req)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error getting room")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}
	resp := room_mapper.GetRoomListToHandlerDto(roomList, req.Limit, req.Offset)

	ctx.JSON(http.StatusOK, resp)
}

// RenameRoom
// @Summary Обновить наименование комнаты
// @Description Обновляет наименование комнаты по ее Id
// @Tags Room
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Id комнаты"
// @Param request body api_dto.UpdateRoomRequest true "Новое наименование комнаты"
// @Success 204 "Данные успешно обновлены"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные данные запроса"
// @Failure 404 {object} middleware_chat.ErrorResponse "Комната не найдена"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room/{id} [put]
func (h *RoomHandler) RenameRoom(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	var req *api_dto.UpdateRoomRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomService.RenameRoomById(ctx, id, req.Name)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error renaming room")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Room renamed successfully"})
}

// DeleteRoom
// @Summary Удалить комнату
// @Description Удаляет комнату по ее Id
// @Tags Room
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Id комнаты"
// @Success 204 "Комната успешно удалена"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные данные запроса"
// @Failure 404 {object} middleware_chat.ErrorResponse "Комната не найдена"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room/{id} [delete]
func (h *RoomHandler) DeleteRoom(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomService.DeleteRoomById(ctx, id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error deleting room")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// GetMemberList
// @Summary Получить список участников комнаты
// @Description Возвращает список участников комнаты по Id комнаты
// @Tags RoomMember
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param room_id path int true "Id комнаты"
// @Success 200 {object} api_dto.GetRoomMemberListResponse "Успешный запрос"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные параметры фильтрации"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room-member/rooms/{room_id}/members [get]
func (h *RoomHandler) GetMemberList(ctx *gin.Context) {
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	memberList, err := h.roomMemberService.ListMembers(ctx, roomId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error getting member list")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}

	resp := room_mapper.GetRoomMemberListToHandlerDto(roomId, memberList)

	ctx.JSON(http.StatusOK, resp)
}

// CreateRoomMember
// @Summary Добавление пользователя в комнату
// @Description Добавляет пользователя в комнату
// @Tags RoomMember
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path int true "Id пользователя"
// @Param room_id path int true "Id комнаты"
// @Success 201 {integer} int "Id созданного члена комнаты"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_chat.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room-member/rooms/{room_id}/members/{user_id} [post]
func (h *RoomHandler) CreateRoomMember(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomMemberService.AddMember(ctx, roomId, userId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error adding member")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusCreated, userId)
}

// SetAdminMember
// @Summary Изменить статус участника комнаты
// @Description Меняет статус участника комнаты на указанный
// @Tags RoomMember
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path int true "Id пользователя"
// @Param room_id path int true "Id комнаты"
// @Param set_admin query boolean true "Назначить администратором (true) или снять (false)"
// @Success 200 {object} map[string]any "Статус участника комнаты успешно изменен"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные данные запроса"
// @Failure 404 {object} middleware_chat.ErrorResponse "Комната не найдена"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room-member/rooms/{room_id}/members/{user_id}/admin [put]
func (h *RoomHandler) SetAdminMember(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	setAdminStr := ctx.Query("set_admin")
	setAdmin, err := strconv.ParseBool(setAdminStr)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "set_admin must be boolean", err), h.log)
		return
	}

	err = h.roomMemberService.SetAdmin(ctx, roomId, userId, setAdmin)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error setting admin")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"message":  "Admin status updated successfully",
		"user_id":  userId,
		"room_id":  roomId,
		"is_admin": setAdmin,
	})
}

// DeleteRoomMember
// @Summary Удалить участника из комнаты
// @Description Удаляет участника с user_id из комнаты с room_id
// @Tags RoomMember
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path int true "Id пользователя"
// @Param room_id path int true "Id комнаты"
// @Success 204 "Пользователь успешно удален из комнаты"
// @Failure 400 {object} middleware_chat.ErrorResponse "Неверные данные запроса"
// @Failure 404 {object} middleware_chat.ErrorResponse "Комната не найдена"
// @Failure 500 {object} middleware_chat.ErrorResponse "Внутренняя ошибка сервера"
// @Router /room-member/rooms/{room_id}/members/{user_id} [delete]
func (h *RoomHandler) DeleteRoomMember(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware_chat.HandleError(ctx, middleware_chat.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomMemberService.RemoveMember(ctx, roomId, userId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error removing member")
		middleware_chat.HandleError(ctx, err, h.log)
		return
	}

	ctx.Status(http.StatusNoContent)
}
