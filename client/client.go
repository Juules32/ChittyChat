package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"

	gRPC "github.com/Juules32/GRPC/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ipFlag string

func main() {
	flag.StringVar(&ipFlag, "ip", "localhost", "IP address (e.g., 192.168.1.1)")
	connectToIP(ipFlag)
}

func connectToIP(IP string) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(IP+":5400", opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	server := gRPC.NewStreamingServiceClient(conn)
	ServerConn := conn
	log.Printf("The connection to %s is: %s\n", IP, conn.GetState().String())
	defer ServerConn.Close()

	stream, err := server.StreamData(context.Background())

	go sendMessage(stream)

	resp, err := stream.Recv()
	if err == io.EOF {
		return // End of the stream
	}
	if err != nil {
		log.Fatalf("Error receiving response: %v", err)
	}

	t1 := resp.TimeSentFromClient
	t2 := resp.TimeReceivedAtServer
	t3 := resp.TimeSentFromServer
	t4 := timestamppb.Now()

	clientDelta := t4.AsTime().Sub(t1.AsTime())
	serverDelta := t3.AsTime().Sub(t2.AsTime())

	delta := clientDelta - serverDelta
	correctedTime := t3.AsTime().Add(delta / 2)
	deltaLocalExpected := t4.AsTime().Sub(correctedTime)
	fmt.Println("t1:", t1.AsTime())
	fmt.Println("t2:", t2.AsTime())
	fmt.Println("t3:", t3.AsTime())
	fmt.Println("t4:", t4.AsTime())
	fmt.Println("total time between request and response:", delta)
	fmt.Printf("Device at IP: %s suggests corrected time: %v\n", IP, correctedTime)
	fmt.Println("Difference between actual and suggested time:", deltaLocalExpected)
	fmt.Println()
}

func sendMessage(stream gRPC.StreamingService_StreamDataClient) {
	t1 := timestamppb.Now()
	// Send data to the server
	greeting := &gRPC.DataRequest{TimeSentFromClient: t1}
	if err := stream.Send(greeting); err != nil {
		log.Fatalf("Error sending data: %v", err)
	}
}
