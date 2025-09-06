package helper

import (
	"chat_service/pkg/grpc_client"
	"context"
	"fmt"
	"proto/generated/profile"
)

func CheckUserExist(ctx context.Context, profileClient *grpc_client.ProfileClient, req *profile.UserExistsRequest) (bool, error) {
	userResp, err := profileClient.UserExists(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return userResp.Exists, nil
}
