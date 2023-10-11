package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	gRPC "github.com/Juules32/GRPC/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	gRPC.UnimplementedStreamingServiceServer        // You need this line if you have a server
	name                                     string // Not required but useful if you want to name your server
	port                                     string // Not required but useful if your server needs to know what port it's listening to

	mutex sync.Mutex // used to lock the server to avoid race conditions.
}

func (s *Server) StreamData(msgStream gRPC.StreamingService_StreamDataServer) error {
	for {
		// get the next message from the stream
		msg, err := msgStream.Recv()
		timeReceivedAtServer := timestamppb.Now()

		fmt.Println("Received Request...")

		time.Sleep(time.Second)
		// the stream is closed so we can exit the loop
		if err == io.EOF {
			return nil
		}
		// some other error
		if err != nil {
			log.Fatalf("Error receiving data: %v", err)
			return err
		}

		fmt.Println("Sending back timestamps...")
		err = msgStream.Send(&gRPC.DataResponse{
			TimeSentFromClient:   msg.TimeSentFromClient,
			TimeReceivedAtServer: timeReceivedAtServer,
			TimeSentFromServer:   timestamppb.Now(),
		})
		if err != nil {
			log.Fatalf("Error sending response: %v", err)
			return err
		}
	}
}

func main() {
	list, err := net.Listen("tcp", ":5400")
	if err != nil {
		return
	}

	grpcServer := grpc.NewServer()

	server := &Server{
		port: "5400",
	}

	gRPC.RegisterStreamingServiceServer(grpcServer, server)

	log.Printf("Server listening at: %v", list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}
