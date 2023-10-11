package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"

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
			fmt.Println("Received message:", res.Message, " From Client:", res.ID)
		}
	}()
	reader := bufio.NewReader(os.Stdin)

	id := rand.Intn(1000000)

	for {
		message, _ := reader.ReadString('\n')

		// Send messages to the server (optional)
		req := &pb.MessageRequest{Message: message, ID: int32(id)}
		if err := stream.Send(req); err != nil {
			log.Printf("Error sending message to server: %v", err)
		}

	}

}
