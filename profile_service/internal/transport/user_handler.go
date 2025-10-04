package transport

import (
	"net/http"
	"os"
	"profile_service/internal/service"
	"profile_service/internal/transport/api_dto"
	"profile_service/internal/transport/user_mapper"
	"profile_service/middleware_profile"
	pb "profile_service/pkg/grpc_generated/profile"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	userService service.UserServiceInterface
	authServer  pb.AuthServiceServer
	log         *logrus.Logger
}

func NewUserHandler(userService service.UserServiceInterface, authServer pb.AuthServiceServer, log *logrus.Logger) *UserHandler {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &UserHandler{
		userService: userService,
		authServer:  authServer,
		log:         log,
	}
}

// PostLogin
// @Summary Вход пользователя
// @Description Аутентифицирует пользователя и возвращает JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body api_dto.LoginRequest true "Данные для входа"
// @Success 200 {object} api_dto.LoginResponse "Успешный вход"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные"
// @Failure 401 {object} middleware_profile.ErrorResponse "Неверные учетные данные"
// @Router /auth/login [post]
func (h *UserHandler) PostLogin(ctx *gin.Context) {
	var req *api_dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid login request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid login request", err), h.log)
		return
	}
	resp, err := h.authServer.Login(ctx.Request.Context(), user_mapper.ConvertToLoginRequest(req))
	if err != nil {
		h.log.WithError(err).Error("Failed to login")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusUnauthorized, "Invalid credentials", err), h.log)
		return
	}
	ctx.JSON(http.StatusOK, user_mapper.ConvertToLoginResponse(resp))
}

// PostUser
// @Summary Создать нового пользователя
// @Description Создает нового пользователя с указанными данными
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body api_dto.CreateUserRequest true "Данные для создания пользователя"
// @Success 201 {integer} int "ID созданного пользователя"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные пользователя"
// @Failure 500 {object} middleware_profile.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func (h *UserHandler) PostUser(ctx *gin.Context) {
	var req *api_dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid create request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid create request", err), h.log)
		return
	}
	mappedReq := user_mapper.ConvertToRegisterRequest(req)
	id, err := h.authServer.Register(ctx.Request.Context(), mappedReq)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to create user")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}
	ctx.JSON(http.StatusCreated, id)
}

// GetUserById
// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по его идентификатору
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} api_dto.UserViewResponse "Успешный запрос"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверный ID пользователя"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} middleware_profile.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/{id} [get]
func (h *UserHandler) GetUserById(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid user ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid user ID", err), h.log)
		return
	}
	user, err := h.userService.GetUserById(ctx.Request.Context(), id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "user_id": id, "path": ctx.Request.URL.Path}).Error("Failed to get user")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}
	resp := user_mapper.ConvertToServiceUser(user)
	ctx.JSON(http.StatusOK, resp)
}

// GetFilteredUserList
// @Summary Получить отфильтрованный список пользователей
// @Description Возвращает список пользователей с возможностью фильтрации
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param name query string false "Фильтр по имени"
// @Param email query string false "Фильтр по email"
// @Param limit query int true "Лимит записей"
// @Param offset query int true "Смещение"
// @Success 200 {object} api_dto.UserViewListResponse "Успешный запрос"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные параметры фильтрации"
// @Failure 500 {object} middleware_profile.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users [get]
func (h *UserHandler) GetFilteredUserList(ctx *gin.Context) {
	var req *api_dto.UserFilterRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid filter request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid filter request", err), h.log)
		return
	}
	filter := user_mapper.ConvertToServiceFilter(req)
	list, err := h.userService.GetAllUsers(ctx.Request.Context(), *filter)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to get users")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}
	resp := user_mapper.ConvertToServiceList(list)
	ctx.JSON(http.StatusOK, resp)
}

// PutUser
// @Summary Обновить данные пользователя
// @Description Обновляет информацию о пользователе по его ID
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param request body api_dto.UpdateUserRequest true "Новые данные пользователя"
// @Success 204 "Данные успешно обновлены"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверные данные запроса"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} middleware_profile.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/{id} [put]
func (h *UserHandler) PutUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid user ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid user ID", err), h.log)
		return
	}
	var req *api_dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Warn("Invalid update request")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid update request", err), h.log)
		return
	}
	mappedReq := user_mapper.ConvertToServiceUpdate(req)
	err = h.userService.UpdateUser(ctx.Request.Context(), id, mappedReq)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to update user")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}
	ctx.JSON(http.StatusNoContent, id)
}

// DeleteUser
// @Summary Удалить пользователя
// @Description Удаляет пользователя по указанному ID
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 204 "Пользователь успешно удален"
// @Failure 400 {object} middleware_profile.ErrorResponse "Неверный ID пользователя"
// @Failure 404 {object} middleware_profile.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} middleware_profile.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Invalid user ID")
		middleware_profile.HandleError(ctx, middleware_profile.NewCustomError(http.StatusBadRequest, "Invalid user ID", err), h.log)
		return
	}
	err = h.userService.DeleteUser(ctx.Request.Context(), id)
	if err != nil {
		h.log.WithFields(logrus.Fields{"error": err, "path": ctx.Request.URL.Path}).Error("Failed to delete user")
		middleware_profile.HandleError(ctx, err, h.log)
		return
	}
	ctx.JSON(http.StatusNoContent, id)
}
