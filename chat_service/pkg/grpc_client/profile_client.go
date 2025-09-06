package grpc_client

import (
	"context"
	"fmt"
	"proto/generated/profile"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProfileClient struct {
	authClient profile.AuthServiceClient
	authConn   *grpc.ClientConn

	userDirClient profile.UserDirectoryClient
	userDirConn   *grpc.ClientConn
}

func NewProfileClient(authAddress, userDirAddress string) (*ProfileClient, error) {
	authConn, err := grpc.NewClient(authAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	authClient := profile.NewAuthServiceClient(authConn)

	userDirConn, err := grpc.NewClient(userDirAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		err = authConn.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	userDirClient := profile.NewUserDirectoryClient(userDirConn)

	return &ProfileClient{
		authClient:    authClient,
		authConn:      authConn,
		userDirClient: userDirClient,
		userDirConn:   userDirConn,
	}, nil
}

func (c *ProfileClient) Close() error {
	var errs []error
	if c.authConn != nil {
		if err := c.authConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.userDirConn != nil {
		if err := c.userDirConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close connections: %v", errs)
	}
	return nil
}

func (c *ProfileClient) ValidateToken(ctx context.Context, req *profile.TokenRequest) (*profile.TokenResponse, error) {
	return c.authClient.ValidateToken(ctx, req)
}

func (c *ProfileClient) UserExists(ctx context.Context, req *profile.UserExistsRequest) (*profile.UserExistsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.userDirClient.UserExists(ctx, req)
}
