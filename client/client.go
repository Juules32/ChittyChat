package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "github.com/Juules32/GRPC/proto" // Import the generated protobuf code
	"google.golang.org/grpc"
)

func main() {
	f, err := os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	var t int32 = 0
	var mu sync.Mutex

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

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.ReplaceAll(name, "\r\n", "")

	// Listen for incoming broadcasts asynchronously
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				break
			}
			mu.Lock()
			tReceived := res.Timestamp
			if tReceived > t {
				t = tReceived
			}
			t++
			log.Println("Client "+name+" receives broadcasted message: \""+res.Message+"\" at timestamp:", +t)
			mu.Unlock()
		}
	}()

	mu.Lock()
	t++
	enterRequest := &pb.Message{Message: name + " has joined the chat", Timestamp: t}
	log.Println("Client "+name+" joins chat at lamport timestamp:", enterRequest.Timestamp)
	if err := stream.Send(enterRequest); err != nil {
		log.Printf("Error sending message to server: %v", err)
	}
	mu.Unlock()

	// Handle termination signal to send a leave message
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh // Wait for termination signal
		mu.Lock()
		t++
		leaveRequest := &pb.Message{Message: name + " has left the chat", Timestamp: t}
		log.Println("Client "+name+" leaves chat at lamport timestamp:", leaveRequest.Timestamp)
		if err := stream.Send(leaveRequest); err != nil {
			log.Printf("Error sending message to server: %v", err)
		}
		mu.Unlock()
		time.Sleep(time.Millisecond * 100)
		os.Exit(0)
	}()

	for {
		fmt.Print("Please enter a message: ")
		message, _ := reader.ReadString('\n')
		message = strings.ReplaceAll(message, "\r\n", "")
		mu.Lock()
		t++
		req := &pb.Message{Message: name + " says: " + message, Timestamp: t}
		log.Println("Client "+name+" publishes message: \""+req.Message+"\" at lamport timestamp:", req.Timestamp)
		if err := stream.Send(req); err != nil {
			log.Printf("Error sending message to server: %v", err)
			return
		}
		mu.Unlock()
		fmt.Println("Message sent successfully!")
	}

}
