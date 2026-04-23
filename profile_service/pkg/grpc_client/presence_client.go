package grpc_client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"profile_service/pkg/grpc_generated/chat"
)

type PresenceClient struct {
	chat.PresenceClient
	conn *grpc.ClientConn
}

func NewPresenceClient(addr string) (*PresenceClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, err
	}

	return &PresenceClient{
		PresenceClient: chat.NewPresenceClient(conn),
		conn:           conn,
	}, nil
}

func (c *PresenceClient) Close() error {
	return c.conn.Close()
}
