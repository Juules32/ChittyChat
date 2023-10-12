package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/Juules32/GRPC/proto" // Import the generated protobuf code
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type chatServer struct {
	clients map[pb.ChatService_SendMessageServer]struct{}
	mu      sync.RWMutex
}

func (s *chatServer) SendMessage(stream pb.ChatService_SendMessageServer) error {

	// Adds client stream to list of clients when connection is established
	s.addClient(stream)

	// Removes client stream from list of clients once connection is cut
	defer s.removeClient(stream)

	// Continuously listens for incoming messages
	for {
		req, err := stream.Recv()
		if err != nil {
			break
		}

		fmt.Println(req.Message)

		// Broadcasts the message to all connected clients
		s.broadcastMessage(req.Message)
	}

	return nil
}

func (s *chatServer) addClient(clientStream pb.ChatService_SendMessageServer) {
	s.mu.Lock()
	s.clients[clientStream] = struct{}{}
	s.mu.Unlock()
}

func (s *chatServer) removeClient(clientStream pb.ChatService_SendMessageServer) {
	s.mu.Lock()
	delete(s.clients, clientStream)
	s.mu.Unlock()
}

func (s *chatServer) broadcastMessage(message string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Broadcast the published message to each client asynchronously
	// This works partially because the response is a stream
	for clientStream := range s.clients {
		go func(stream pb.ChatService_SendMessageServer, message string) {
			if err := stream.Send(&pb.Message{Message: message}); err != nil {
				log.Printf("Error sending message to client: %v", err)
			}
		}(clientStream, message)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, &chatServer{
		clients: make(map[pb.ChatService_SendMessageServer]struct{}),
	})
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
}
