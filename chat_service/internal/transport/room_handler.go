package transport

import (
	"chat_service/internal/service"
	"chat_service/internal/transport/api_dto"
	"chat_service/internal/transport/room_mapper"
	"chat_service/middleware"
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

func (h *RoomHandler) CreateRoom(ctx *gin.Context) {
	var req *api_dto.CreateRoomRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	id, err := h.roomService.CreateRoom(ctx, req.Name)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error creating room")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusCreated, id)
}

func (h *RoomHandler) GetRoom(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	room, err := h.roomService.GetRoomById(ctx, id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error getting room")
		middleware.HandleError(ctx, err, h.log)
		return
	}
	resp := room_mapper.GetRoomToHandlerDto(room)

	ctx.JSON(http.StatusOK, resp)

}

func (h *RoomHandler) GetRoomList(ctx *gin.Context) {
	var searchFilter *api_dto.SearchRoomRequest
	if err := ctx.ShouldBindQuery(&searchFilter); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	req := room_mapper.SearchQueryToServiceFilter(searchFilter)
	roomList, err := h.roomService.GetRoomList(ctx, req)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error getting room")
		middleware.HandleError(ctx, err, h.log)
		return
	}
	resp := room_mapper.GetRoomListToHandlerDto(roomList, req.Limit, req.Offset)

	ctx.JSON(http.StatusOK, resp)
}

func (h *RoomHandler) RenameRoom(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err = ctx.ShouldBindJSON(&id); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	var req *api_dto.UpdateRoomRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomService.RenameRoomById(ctx, id, req.Name)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error renaming room")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusCreated, id)
}

func (h *RoomHandler) DeleteRoom(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomService.DeleteRoomById(ctx, id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error deleting room")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (h *RoomHandler) CreateRoomMember(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomMemberService.AddMember(ctx, userId, roomId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error adding member")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusCreated, userId)
}

func (h *RoomHandler) DeleteRoomMember(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	err = h.roomMemberService.RemoveMember(ctx, userId, roomId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error removing member")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (h *RoomHandler) SetAdminMember(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}
	var req *api_dto.SetAdminStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{
			"error":   err,
			"user_id": userId,
			"room_id": roomId,
		}).Warn("Invalid request body")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request body", err), h.log)
		return
	}

	err = h.roomMemberService.SetAdmin(ctx, userId, roomId, req.SetAdmin)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error setting admin")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	resp := gin.H{
		"user_id":  userId,
		"room_id":  roomId,
		"is_admin": req.SetAdmin,
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *RoomHandler) GetMemberList(ctx *gin.Context) {
	roomId, err := strconv.ParseInt(ctx.Param("room_id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	memberList, err := h.roomMemberService.ListMembers(ctx, roomId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error getting member list")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	resp := room_mapper.GetRoomMemberListToHandlerDto(roomId, memberList)

	ctx.JSON(http.StatusOK, resp)
}
