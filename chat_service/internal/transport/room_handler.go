package transport

import (
	"chat_service/internal/service"
	"chat_service/internal/transport/api_dto"
	"chat_service/internal/transport/room_mapper"
	"net/http"
	"os"
	"proto/middleware"
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

}

func (h *RoomHandler) DeleteRoomMember(ctx *gin.Context) {

}

func (h *RoomHandler) SetAdminMember(ctx *gin.Context) {

}

func (h *RoomHandler) GetMemberList(ctx *gin.Context) {

}
