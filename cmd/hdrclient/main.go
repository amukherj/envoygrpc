package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/amukherj/envoygrpc/messages"
	"github.com/amukherj/envoygrpc/names"
)

const (
	routingHeader = "x-ikat-service-id"
)

var address string = "localhost:50501"

func main() {
	msg := "Go rules!"
	if len(os.Args) <= 1 {
		log.Fatalf("usage: %s <server-IP:port> [[msg] header value]",
			os.Args[0])
	}
	address = os.Args[1]

	if len(os.Args) > 2 {
		msg = os.Args[2]
	}

	var header, headerVal string
	if len(os.Args) > 4 {
		header = os.Args[3]
		headerVal = os.Args[4]
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
	}
	// grpc.WithInsecure()

	if header == "authority" {
		dialOptions = []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithAuthority(headerVal),
			grpc.WithBlock(),
		}
		header = ""
	}

	conn, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		log.Fatalf("[client] Failed to connect to grpc server: %v", err)
	}
	defer conn.Close()

	// Call the EchoService
	client := messages.NewEchoServiceClient(conn)

	serverName, _ := os.Hostname()
	now := time.Now().Unix()
	payload := messages.EchoMessage{
		ServerName: &serverName,
		UtcTime:    &now,
		Msg:        &msg,
	}

	var headers metadata.MD
	if len(header) > 0 {
		if header == routingHeader && headerVal == "text" {
			headers = metadata.Pairs(header, "greeter")
		} else {
			headers = metadata.Pairs(header, headerVal)
		}
	}

	baseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ctx := metadata.NewOutgoingContext(baseCtx, headers)

	log.Printf("[client] initiating gRPC request with headers")
	resp, err := client.Hello(ctx, &payload)
	if err != nil {
		log.Printf("[client] RPC error: %v", err)
	} else {
		log.Printf(`[client] Response message:
	From: %s
	Sent-at: %d
	Msg: %s`, resp.GetServerName(), resp.GetUtcTime(), resp.GetMsg())
	}

	// Call the GreatNamesService
	client1 := names.NewGreatNamesServiceClient(conn)
	now = time.Now().Unix()
	request := &names.NameRequest{
		ServerName: &serverName,
		UtcTime:    &now,
	}

	if len(header) > 0 {
		if header == routingHeader && headerVal == "greeter" {
			headers = metadata.Pairs(header, "text")
		} else {
			headers = metadata.Pairs(header, headerVal)
		}
	}
	ctx = metadata.NewOutgoingContext(baseCtx, headers)
	resp1, err := client1.Get(ctx, request)
	if err != nil {
		log.Fatalf("[client] RPC error: %v", err)
	}
	log.Printf(`[client] Response message:
	From: %s
	Sent-at: %d
	Response: %s %s`, resp1.GetServerName(), resp1.GetUtcTime(),
		resp1.GetQuality(), resp1.GetPerson())
}
