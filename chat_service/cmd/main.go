package main

import (
	"chat_service/pkg/grpc_client"
	"log"
)

func main() {
	profileClient, err := grpc_client.NewProfileClient("profileService:50051", "profileService:50052")
	if err != nil {
		log.Fatalf("failed to create profile client: %v", err)
	}
	defer func() {
		if err := profileClient.Close(); err != nil {
			log.Printf("failed to close profile client: %v", err)
		}
	}()
}
