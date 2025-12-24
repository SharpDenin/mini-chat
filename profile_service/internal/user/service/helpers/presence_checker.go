package helpers

import (
	"context"
	"errors"
	"fmt"
	"profile_service/pkg/grpc_client"
	"profile_service/pkg/grpc_generated/chat"
	"time"
)

func CheckUserPresence(ctx context.Context, presenceClient *grpc_client.PresenceClient, req *chat.GetPresenceRequest) (bool, error) {
	if presenceClient == nil {
		return false, errors.New("presence client is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status, err := presenceClient.GetPresence(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to get presence for user %v: %w", req.UserId, err)
	}

	if status == nil {
		return false, nil
	}

	presence := status.GetPresence()
	if presence == nil {
		return false, nil
	}

	switch presence.Status {
	case chat.UserStatus_ONLINE:
		return true, nil
	case chat.UserStatus_IDLE:
		return true, nil
	case chat.UserStatus_OFFLINE:
		return false, nil
	default:
		return false, nil
	}
}
