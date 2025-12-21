package grpc_client

import (
	"chat_service/pkg/grpc_generated/chat"
	"fmt"

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
