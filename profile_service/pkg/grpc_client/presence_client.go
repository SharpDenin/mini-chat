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
	presenceClient chat.PresenceServiceClient
	presenceConn   *grpc.ClientConn
}

func NewPresenceClient(presenceAddress string) (*PresenceClient, error) {
	presenceConn, err := grpc.Dial(presenceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("fail to dial presence service: %v", err)
	}
	presenceClient := chat.NewPresenceServiceClient(presenceConn)

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

func (pc *PresenceClient) GetRecentlyOnline(ctx context.Context, req *chat.GetRecentlyOnlineRequest, opt ...grpc.CallOption) (*chat.GetRecentlyOnlineResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return pc.presenceClient.GetRecentlyOnline(ctx, req, opt...)
}

func (pc *PresenceClient) GetPresence(ctx context.Context, req *chat.GetPresenceRequest) (*chat.GetPresenceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pc.presenceClient.GetPresence(ctx, req)
}

func (pc *PresenceClient) GetBulkPresence(ctx context.Context, req *chat.GetBulkPresenceRequest) (*chat.GetBulkPresenceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pc.presenceClient.GetBulkPresence(ctx, req)
}

func (pc *PresenceClient) MarkOnline(ctx context.Context, req *chat.MarkOnlineRequest, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := pc.presenceClient.MarkOnline(ctx, req, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (pc *PresenceClient) MarkOffline(ctx context.Context, req *chat.MarkOfflineRequest, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := pc.presenceClient.MarkOffline(ctx, req, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (pc *PresenceClient) UpdateLastSeen(ctx context.Context, req *chat.UpdateLastSeenRequest, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := pc.presenceClient.UpdateLastSeen(ctx, req, opts...)
	if err != nil {
		return err
	}
	return nil
}
