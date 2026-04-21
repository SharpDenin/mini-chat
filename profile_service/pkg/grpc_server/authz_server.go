package grpc_server

import (
	"context"
	"profile_service/internal/relation/service/interfaces"
	"profile_service/pkg/grpc_generated/profile"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthorizationServer struct {
	profile.UnimplementedAuthorizationServiceServer
	relationChecker interfaces.UserRelationCheckerInterface
}

func NewAuthorizationServer(svc interfaces.UserRelationCheckerInterface) *AuthorizationServer {
	return &AuthorizationServer{
		relationChecker: svc,
	}
}

func (s *AuthorizationServer) CanSendDirect(ctx context.Context, req *profile.CanSendDirectRequest) (*profile.CanSendDirectResponse, error) {
	if req.FromUserId == req.ToUserId {
		return &profile.CanSendDirectResponse{Allowed: true}, nil
	}

	blocked, err := s.relationChecker.CheckUserIsBlocked(ctx, req.ToUserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "block check failed")
	}
	if blocked {
		return &profile.CanSendDirectResponse{
			Allowed: false,
			Reason:  "blocked",
		}, nil
	}

	friends, err := s.relationChecker.CheckUsersAreFriends(ctx, req.FromUserId, req.ToUserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "friends check failed")
	}
	if !friends {
		return &profile.CanSendDirectResponse{
			Allowed: false,
			Reason:  "not_friends",
		}, nil
	}

	return &profile.CanSendDirectResponse{
		Allowed: true,
	}, nil
}

func (s *AuthorizationServer) CanJoinRoom(ctx context.Context, req *profile.CanJoinRoomRequest) (*profile.CanJoinRoomResponse, error) {
	return &profile.CanJoinRoomResponse{
		Allowed: true,
	}, nil
}
