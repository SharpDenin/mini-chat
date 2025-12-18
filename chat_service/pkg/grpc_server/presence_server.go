package grpc_server

import (
	"chat_service/internal/presence/service"
	dto "chat_service/internal/presence/service/dto"
	"chat_service/pkg/grpc_generated/chat"
	"context"
	"errors"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PresenceServer struct {
	log *logrus.Logger
	chat.UnimplementedPresenceServiceServer
	pService service.PresenceServiceInterface
}

func NewPresenceServer(log *logrus.Logger, pService service.PresenceServiceInterface) *PresenceServer {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &PresenceServer{
		log:      log,
		pService: pService,
	}
}

func (s *PresenceServer) MarkOnline(ctx context.Context, req *chat.MarkOnlineRequest) (*emptypb.Empty, error) {
	s.log.WithFields(logrus.Fields{
		"method":  "MarkOnline",
		"user_id": req.UserId,
		"source":  req.Source,
	}).Debug("MarkOnline request")

	if req.GetUserId() == 0 {
		s.log.WithField("error", "user_id is required").Warn("Validation failed")
		return nil, status.Error(codes.InvalidArgument, "UserId is required")
	}

	var opts []dto.MarkOptionRequest
	if source := req.GetSource(); source != "" {
		opts = append(opts, dto.WithSource(source))
	}

	if err := s.pService.MarkOnline(ctx, req.GetUserId(), opts...); err != nil {
		s.log.WithFields(logrus.Fields{
			"error": err.Error(),
			"opts":  opts,
		}).Error("MarkOnline failed")
		return nil, s.mapServiceErrorToGRPC(err)
	}

	s.log.Info("User marked online")

	return &emptypb.Empty{}, nil
}

func (s *PresenceServer) MarkOffline(ctx context.Context, req *chat.MarkOfflineRequest) (*emptypb.Empty, error) {
	s.log.WithFields(logrus.Fields{
		"method":  "MarkOffline",
		"user_id": req.UserId,
		"source":  req.Source,
	}).Debug("MarkOffline request")

	if req.GetUserId() == 0 {
		s.log.WithField("error", "user_id is required").Warn("Validation failed")
		return nil, status.Error(codes.InvalidArgument, "UserId is required")
	}

	var opts []dto.MarkOptionRequest
	if source := req.GetSource(); source != "" {
		opts = append(opts, dto.WithSource(source))
	}

	if err := s.pService.MarkOffline(ctx, req.GetUserId(), opts...); err != nil {
		s.log.WithFields(logrus.Fields{
			"error": err.Error(),
			"opts":  opts,
		}).Error("MarkOffline failed")
		return nil, s.mapServiceErrorToGRPC(err)
	}

	s.log.Info("User marked offline")

	return &emptypb.Empty{}, nil
}

func (s *PresenceServer) UpdateLastSeen(ctx context.Context, req *chat.UpdateLastSeenRequest) (*emptypb.Empty, error) {
	s.log.WithFields(logrus.Fields{
		"method":  "UpdateLastSeen",
		"user_id": req.UserId,
	}).Debug("UpdateLastSeen request")

	if req.GetUserId() == 0 {
		s.log.WithField("error", "user_id is required").Warn("Validation failed")
		return nil, status.Error(codes.InvalidArgument, "User_id is required")
	}

	err := s.pService.UpdateLastSeen(ctx, req.GetUserId())
	if err != nil {
		s.log.WithError(err).Error("UpdateLastSeen service call failed")
		return nil, s.mapServiceErrorToGRPC(err)
	}

	s.log.Info("Last seen updated successfully")

	return &emptypb.Empty{}, nil
}

func (s *PresenceServer) GetPresence(ctx context.Context, req *chat.GetPresenceRequest) (*chat.GetPresenceResponse, error) {
	s.log.WithFields(logrus.Fields{
		"method":  "GetPresence",
		"user_id": req.UserId,
	}).Debug("GetPresence request")

	if req.GetUserId() == 0 {
		s.log.WithField("error", "user_id is required").Warn("Validation failed")
		return nil, status.Error(codes.InvalidArgument, "UserId is required")
	}

	presence, err := s.pService.GetPresence(ctx, req.GetUserId())
	if err != nil {
		s.log.WithError(err).Error("GetPresence service call failed")
		return nil, s.mapServiceErrorToGRPC(err)
	}

	pbPresence := s.convertDomainToProto(presence)

	s.log.WithFields(logrus.Fields{
		"status":    pbPresence.GetStatus(),
		"last_seen": pbPresence.GetLastSeen(),
	}).Debug("GetPresence completed")

	return &chat.GetPresenceResponse{
		Presence: pbPresence,
	}, nil
}

func (s *PresenceServer) GetBulkPresence(ctx context.Context, req *chat.GetBulkPresenceRequest) (*chat.GetBulkPresenceResponse, error) {
	s.log.WithFields(logrus.Fields{
		"method":     "GetBulkPresence",
		"user_count": len(req.GetUser_Ids()),
	}).Debug("GetBulkPresence request")

	if len(req.GetUser_Ids()) == 0 {
		s.log.Warn("User_id is required")
		return &chat.GetBulkPresenceResponse{
			Presences: make(map[int64]*chat.Presence),
		}, nil
	}

	userIds := req.GetUser_Ids()
	if limit := req.GetLimit(); limit > 0 && len(userIds) > int(limit) {
		userIds = userIds[:limit]
		s.log.WithField("truncated_to", limit).Warn("User list truncated")
	}

	bulkResp, err := s.pService.GetBulkPresence(ctx, userIds)
	if err != nil {
		s.log.WithError(err).Error("GetBulkPresence service call failed")
		return nil, s.mapServiceErrorToGRPC(err)
	}

	resp := &chat.GetBulkPresenceResponse{
		Presences: make(map[int64]*chat.Presence),
		Errors:    make(map[int64]string),
	}

	for userId, presence := range bulkResp.Presences {
		resp.Presences[userId] = s.convertDomainToProto(presence)
	}

	s.log.WithFields(logrus.Fields{
		"success_count": len(resp.Presences),
		"error_count":   len(resp.Errors),
	}).Info("GetBulkPresence completed")

	return resp, nil
}

func (s *PresenceServer) GetOnlineUsers(ctx context.Context, req *chat.GetOnlineUsersRequest) (*chat.GetOnlineUsersResponse, error) {
	s.log.WithFields(logrus.Fields{
		"method":     "GetOnlineUsers",
		"user_count": len(req.GetUserIds()),
	}).Debug("GetOnlineUsers request")

	if len(req.GetUserIds()) == 0 {
		s.log.Warn("User_id list is empty")
		return &chat.GetOnlineUsersResponse{
			OnlineUserIds: []int64{},
		}, nil
	}

	onlineUsers, err := s.pService.GetOnlineUsers(ctx, req.GetUserIds())
	if err != nil {
		s.log.WithError(err).Error("GetOnlineUsers service call failed")
		return nil, s.mapServiceErrorToGRPC(err)
	}

	logrus.WithFields(logrus.Fields{
		"online_count": len(onlineUsers),
	}).Info("GetOnlineUsers completed")

	return &chat.GetOnlineUsersResponse{
		OnlineUserIds: onlineUsers,
	}, nil
}

// GetRecentlyOnline
// TODO Доработки
func (s *PresenceServer) GetRecentlyOnline(ctx context.Context, req *chat.GetRecentlyOnlineRequest) (*chat.GetRecentlyOnlineResponse, error) {
	s.log.WithFields(logrus.Fields{
		"method": "GetRecentlyOnline",
		"since":  req.GetSince(),
		"limit":  req.GetLimit(),
	}).Debug("GetRecentlyOnline request")

	var sinceTime time.Time
	if ts := req.GetSince(); ts != nil {
		sinceTime = ts.AsTime()
	} else {
		sinceTime = time.Now().Add(24 * time.Hour)
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 100
	}

	recentlyOnline, err := s.pService.GetRecentlyOnline(ctx, sinceTime)
	if err != nil {
		s.log.WithError(err).Error("GetRecentlyOnline service call failed")
		return nil, s.mapServiceErrorToGRPC(err)
	}

	if len(recentlyOnline) > int(limit) {
		recentlyOnline = recentlyOnline[:limit]
	}

	users := make([]*chat.RecentlyOnlineUser, 0, len(recentlyOnline))
	for i, userId := range recentlyOnline {
		users = append(users, &chat.RecentlyOnlineUser{
			UserId:     userId,
			LastSeen:   users[i].LastSeen,
			LastStatus: users[i].LastStatus,
		})
	}

	s.log.WithFields(logrus.Fields{
		"user_count": len(users),
	}).Info("GetRecentlyOnline completed")

	return &chat.GetRecentlyOnlineResponse{
		Users: users,
	}, nil
}

// HealthCheck
//TODO Доработки
func (s *PresenceServer) HealthCheck(ctx context.Context, req *emptypb.Empty) (*chat.HealthCheckResponse, error) {
	s.log.Info("Health check request")
	return &chat.HealthCheckResponse{
		Healthy:       true,
		Status:        "ok",
		TimestampUnix: time.Now().Unix(),
	}, nil
}

func (s *PresenceServer) convertDomainToProto(presence *dto.PresenceResponse) *chat.Presence {
	if presence == nil {
		return nil
	}

	var lastSeenPb *timestamppb.Timestamp
	if !presence.LastSeen.IsZero() {
		lastSeenPb = timestamppb.New(presence.LastSeen)
	} else {
		lastSeenPb = timestamppb.New(time.Now())
	}

	pbPresence := &chat.Presence{
		UserId:   presence.UserId,
		LastSeen: lastSeenPb,
	}

	if presence.DeviceType != "" {
		pbPresence.DeviceType = &presence.DeviceType
	}

	isIdle := presence.Status == dto.StatusIdle
	pbPresence.IsIdle = &isIdle

	switch presence.Status {
	case dto.StatusOnline:
		pbPresence.Status = chat.UserStatus_ONLINE
	case dto.StatusOffline:
		pbPresence.Status = chat.UserStatus_OFFLINE
	case dto.StatusIdle:
		pbPresence.Status = chat.UserStatus_IDLE
		isIdle = true
		pbPresence.IsIdle = &isIdle
	default:
		pbPresence.Status = chat.UserStatus_UNKNOWN
	}

	return pbPresence
}

func (s *PresenceServer) mapServiceErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, dto.ErrInvalidUserId):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, dto.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, dto.ErrRateLimitExceeded):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, dto.ErrAlreadyOnline), errors.Is(err, dto.ErrAlreadyOffline):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, dto.ErrServiceUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "request timeout")
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "request cancelled")
	}

	s.log.WithError(err).Error("Unmapped service error")

	return status.Error(codes.Internal, "internal server error")
}
