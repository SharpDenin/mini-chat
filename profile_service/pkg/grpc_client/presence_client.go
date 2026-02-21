package grpc_client

import (
	"context"
	"fmt"
	"profile_service/pkg/grpc_generated/chat"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PresenceClient struct {
	presenceClient chat.PresenceClient
	presenceConn   *grpc.ClientConn
}

func NewPresenceClient(presenceAddress string) (*PresenceClient, error) {
	presenceConn, err := grpc.Dial(presenceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("fail to dial presence service: %v", err)
	}
	presenceClient := chat.NewPresenceClient(presenceConn)

	return &PresenceClient{
		presenceClient: presenceClient,
		presenceConn:   presenceConn,
	}, nil
}

func (pc *PresenceClient) Close() error {
	var errs []error
	if pc.presenceConn != nil {
		if err := pc.presenceConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("fail to close presence client: %v", errs)
	}

	return nil
}

func (pc *PresenceClient) OnConnect(ctx context.Context, req *chat.OnConnectRequest, opts ...grpc.CallOption) (*chat.EmptyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pc.presenceClient.OnConnect(ctx, req)
}

func (pc *PresenceClient) OnDisconnect(ctx context.Context, req *chat.OnDisconnectRequest, opts ...grpc.CallOption) (*chat.EmptyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pc.presenceClient.OnDisconnect(ctx, req)
}

func (pc *PresenceClient) OnHeartbeat(ctx context.Context, req *chat.OnHeartbeatRequest, opts ...grpc.CallOption) (*chat.EmptyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pc.presenceClient.OnHeartbeat(ctx, req)
}

func (pc *PresenceClient) GetPresence(ctx context.Context, req *chat.GetPresenceRequest) (*chat.GetPresenceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pc.presenceClient.GetPresence(ctx, req)
}

func (pc *PresenceClient) GetOnlineFriends(ctx context.Context, req *chat.GetOnlineFriendsRequest, opts ...grpc.CallOption) (*chat.GetOnlineFriendsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pc.presenceClient.GetOnlineFriends(ctx, req)
}
