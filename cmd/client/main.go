package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/amukherj/envoygrpc/messages"
)

const (
	address = "localhost:50501"
)

func main() {
	msg := "Go rules!"
	if len(os.Args) > 1 {
		msg = os.Args[1]
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to grpc server: %v", err)
	}
	defer conn.Close()

	client := messages.NewEchoServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	serverName, _ := os.Hostname()
	now := time.Now().Unix()
	payload := messages.EchoMessage{
		ServerName: &serverName,
		UtcTime:    &now,
		Msg:        &msg,
	}
	resp, err := client.Hello(ctx, &payload)
	if err != nil {
		log.Fatalf("RPC error: %v", err)
	}
	log.Printf(`Response message:
	From: %s
	Sent-at: %d
	Msg: %s`, resp.GetServerName(), resp.GetUtcTime(), resp.GetMsg())
}
