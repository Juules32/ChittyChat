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
	clients map[chan<- *pb.MessageResponse]struct{}
	mu      sync.RWMutex
}

func newChatServer() *chatServer {
	return &chatServer{
		clients: make(map[chan<- *pb.MessageResponse]struct{}),
	}
}

func (s *chatServer) SendMessage(stream pb.ChatService_SendMessageServer) error {
	clientCh := make(chan *pb.MessageResponse)
	s.addClient(clientCh)
	defer s.removeClient(clientCh)

	for {
		msgReq, err := stream.Recv()

		if err != nil {

		}
		fmt.Println(msgReq.Message)

		// Broadcast the message to all connected clients
		s.broadcastMessage(msgReq.Message)

		// Send a response to the client (optional)
		response := &pb.MessageResponse{Message: "Message received"}
		if err := stream.Send(response); err != nil {
			log.Printf("Error sending response to client: %v", err)
		}
	}

	return nil
}

func (s *chatServer) addClient(clientCh chan<- *pb.MessageResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[clientCh] = struct{}{}
}

func (s *chatServer) removeClient(clientCh chan<- *pb.MessageResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, clientCh)
	close(clientCh)
}

func (s *chatServer) broadcastMessage(message string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for clientCh := range s.clients {
		fmt.Println("heya")
		go func(ch chan<- *pb.MessageResponse, message string) {
			ch <- &pb.MessageResponse{Message: message}
		}(clientCh, message)
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
