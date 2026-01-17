package grpc_server

import (
	"chat_service/internal/presence/service"
	"chat_service/pkg/grpc_generated/chat"
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCServer struct {
	svc service.PresenceService
	chat.UnimplementedPresenceServer
}

func NewGRPCServer(svc service.PresenceService) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) OnConnect(ctx context.Context, req *chat.OnConnectRequest) (*chat.EmptyResponse, error) {
	if err := s.svc.OnConnect(ctx, req.UserId, req.ConnId, req.Device); err != nil {
		return nil, err
	}
	return &chat.EmptyResponse{}, nil
}

func (s *GRPCServer) OnDisconnect(ctx context.Context, req *chat.OnDisconnectRequest) (*chat.EmptyResponse, error) {
	if err := s.svc.OnDisconnect(ctx, req.UserId, req.ConnId); err != nil {
		return nil, err
	}
	return &chat.EmptyResponse{}, nil
}

func (s *GRPCServer) OnHeartbeat(ctx context.Context, req *chat.OnHeartbeatRequest) (*chat.EmptyResponse, error) {
	if err := s.svc.OnHeartbeat(ctx, req.ConnId); err != nil {
		return nil, err
	}
	return &chat.EmptyResponse{}, nil
}

func (s *GRPCServer) GetPresence(ctx context.Context, req *chat.GetPresenceRequest) (*chat.GetPresenceResponse, error) {
	p := s.svc.GetPresence(ctx, req.UserId)
	return &chat.GetPresenceResponse{
		UserId:   p.UserId,
		Status:   string(p.Status),
		LastSeen: timestamppb.New(p.LastSeen),
	}, nil
}

func (s *GRPCServer) GetOnlineFriends(ctx context.Context, req *chat.GetOnlineFriendsRequest) (*chat.GetOnlineFriendsResponse, error) {
	online := s.svc.GetOnlineFriends(ctx, req.UserId, req.FriendsIds)
	return &chat.GetOnlineFriendsResponse{
		OnlineFriends: online,
	}, nil
}
