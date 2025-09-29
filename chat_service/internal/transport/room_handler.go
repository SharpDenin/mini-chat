package transport

import (
	"chat_service/internal/service"
	"net/http"
	"os"
	"proto/middleware"

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
	var name string
	if err := ctx.ShouldBindJSON(&name); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid request parameters")
		middleware.HandleError(ctx, middleware.NewCustomError(http.StatusBadRequest, "Invalid request parameters", err), h.log)
		return
	}

	id, err := h.roomService.CreateRoom(ctx, name)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Error creating room")
		middleware.HandleError(ctx, err, h.log)
		return
	}

	ctx.JSON(http.StatusCreated, id)
}

func (h *RoomHandler) GetRoom(ctx *gin.Context) {
	//TODO Проверяем входные параметры и инициализируем DTO request-а (api_dto) и биндим из контекста
	//TODO Конвертируем в service_dto и вызываем метод сервиса
	//TODO Ответ в формате ```ctx.JSON(http.StatusCreated, id)```
}

func (h *RoomHandler) GetRoomList(ctx *gin.Context) {

}

func (h *RoomHandler) RenameRoom(ctx *gin.Context) {

}

func (h *RoomHandler) DeleteRoom(ctx *gin.Context) {

}

func (h *RoomHandler) CreateRoomMember(ctx *gin.Context) {

}

func (h *RoomHandler) DeleteRoomMember(ctx *gin.Context) {

}

func (h *RoomHandler) SetAdminMember(ctx *gin.Context) {

}

func (h *RoomHandler) GetMemberList(ctx *gin.Context) {

}
