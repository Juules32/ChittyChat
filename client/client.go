package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	// Listen for incoming broadcasts asynchronously
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
			}
			fmt.Println(res.Message)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.ReplaceAll(name, "\r\n", "")

	enterRequest := &pb.Message{Message: name + " has joined the chat"}
	if err := stream.Send(enterRequest); err != nil {
		log.Printf("Error sending message to server: %v", err)
	}

	// Handle termination signal to send a leave message
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh // Wait for termination signal
		leaveRequest := &pb.Message{Message: name + " has left the chat"}
		if err := stream.Send(leaveRequest); err != nil {
			log.Printf("Error sending message to server: %v", err)
		}
		time.Sleep(time.Millisecond * 100)
		os.Exit(0)
	}()

	for {
		message, _ := reader.ReadString('\n')
		message = strings.ReplaceAll(message, "\r\n", "")

		req := &pb.Message{Message: name + " says: " + message}
		if err := stream.Send(req); err != nil {
			log.Printf("Error sending message to server: %v", err)
		}

	}

}
