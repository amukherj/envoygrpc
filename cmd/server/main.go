package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/amukherj/envoygrpc/messages"
)

type MsgService struct{}

func (m MsgService) Hello(ctx context.Context,
	in *messages.EchoMessage) (*messages.EchoMessage, error) {

	log.Printf(`Message received: \n
	From: %s
	Sent-at: %d
	Content: %s`,
		in.GetServerName(), in.GetUtcTime(), in.GetMsg())

	hostname, _ := os.Hostname()
	response := "Response: " + in.GetMsg()
	now := time.Now().Unix()
	in.ServerName = &hostname
	in.UtcTime = &now
	in.Msg = &response

	return in, nil
}

func main() {
	port := ":50501"
	if len(os.Args) > 1 {
		port = os.Args[1]
		log.Println("len(os.Args) > 1: ", port)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to start listening on %s: %v", port, err)
	}

	svrName, err := os.Hostname()
	svr := grpc.NewServer()
	var svc MsgService
	messages.RegisterEchoServiceServer(svr, svc)
	log.Printf("Starting GRPC server on %s port %s ...\n", svrName, port)
	if err := svr.Serve(lis); err != nil {
		log.Fatalf("Failed to start GRPC service: %v", err)
	}
}
