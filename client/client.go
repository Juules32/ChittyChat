package main

import (
	"context"
	"log"

	pb "github.com/Juules32/GRPC/proto" // Import the generated protobuf code
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)

	stream, err := client.SendMessage(context.Background())
	if err != nil {
		log.Fatalf("Error creating stream: %v", err)
	}

	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
			}
			log.Printf("Received message from server: %s", res.Message)
		}
	}()

	// Send messages to the server (optional)
	messages := []string{"Hello", "World"}
	for _, message := range messages {
		req := &pb.MessageRequest{Message: message}
		if err := stream.Send(req); err != nil {
			log.Printf("Error sending message to server: %v", err)
			break
		}
	}

	for {

	}

}
