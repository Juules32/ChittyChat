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

func newChatServer() *chatServer {
	return &chatServer{
		clients: make(map[pb.ChatService_SendMessageServer]struct{}),
	}
}

func (s *chatServer) SendMessage(stream pb.ChatService_SendMessageServer) error {
	s.addClient(stream)
	defer s.removeClient(stream)

	for {
		msgReq, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving message from client: %v", err)
			break
		}

		fmt.Println("Received message:", msgReq.Message, "From client:", msgReq.ID)

		// Broadcast the message to all connected clients
		s.broadcastMessage(msgReq.Message, msgReq.ID)
	}

	return nil
}

func (s *chatServer) addClient(clientStream pb.ChatService_SendMessageServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[clientStream] = struct{}{}
}

func (s *chatServer) removeClient(clientStream pb.ChatService_SendMessageServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, clientStream)
}

func (s *chatServer) broadcastMessage(message string, id int32) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for clientStream := range s.clients {
		go func(stream pb.ChatService_SendMessageServer, message string) {
			response := &pb.MessageResponse{Message: message, ID: id}
			if err := stream.Send(response); err != nil {
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
	pb.RegisterChatServiceServer(grpcServer, newChatServer())
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
