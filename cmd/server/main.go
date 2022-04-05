package main

import (
	"context"
	"io"
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

	log.Printf(`[echo] Message received: \n
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

func (m MsgService) WhatsUp(stream messages.EchoService_WhatsUpServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf(`[echo] Request received: \n
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
		if err = stream.Send(in); err != nil {
			return err
		}
	}
}

func main() {
	port := ":50501"
	if len(os.Args) > 1 {
		port = os.Args[1]
		log.Println("len(os.Args) > 1: ", port)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("[echo] Failed to start listening on %s: %v", port, err)
	}

	svrName, err := os.Hostname()
	svr := grpc.NewServer()
	var svc MsgService
	messages.RegisterEchoServiceServer(svr, svc)
	log.Printf("[echo] Starting GRPC server on %s port %s ...\n", svrName, port)
	if err := svr.Serve(lis); err != nil {
		log.Fatalf("[echo] Failed to start GRPC service: %v", err)
	}
}
