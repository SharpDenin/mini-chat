package authz

import (
	"chat_service/pkg/grpc_client"
	"context"
)

type GrpcAuthz struct {
	profileClient *grpc_client.ProfileClient
}

func NewGrpcAuthz(client *grpc_client.ProfileClient) *GrpcAuthz {
	return &GrpcAuthz{
		profileClient: client,
	}
}

func (a *GrpcAuthz) CanSendDirect(ctx context.Context, fromUserId, toUserId int64) (bool, error) {
	resp, err := a.profileClient.CanSendDirect(ctx, fromUserId, toUserId)
	if err != nil {
		return false, err
	}

	return resp.Allowed, nil
}
