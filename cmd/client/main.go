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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	serverName, _ := os.Hostname()
	now := time.Now().Unix()
	payload := messages.EchoMessage{
		ServerName: &serverName,
		UtcTime:    &now,
		Msg:        &msg,
	}

	headers := metadata.MD{}
	if len(header) > 0 {
		headers[header] = []string{headerVal}
	}

	resp, err := client.Hello(ctx, &payload, grpc.Header(&headers))
	if err != nil {
		log.Fatalf("[client] RPC error: %v", err)
	}
	log.Printf(`[client] Response message:
	From: %s
	Sent-at: %d
	Msg: %s`, resp.GetServerName(), resp.GetUtcTime(), resp.GetMsg())

	// Call the GreatNamesService
	client1 := names.NewGreatNamesServiceClient(conn)
	now = time.Now().Unix()
	request := &names.NameRequest{
		ServerName: &serverName,
		UtcTime:    &now,
	}

	resp1, err := client1.Get(ctx, request, grpc.Header(&headers))
	if err != nil {
		log.Fatalf("[client] RPC error: %v", err)
	}
	log.Printf(`[client] Response message:
	From: %s
	Sent-at: %d
	Response: %s %s`, resp1.GetServerName(), resp1.GetUtcTime(),
		resp1.GetQuality(), resp1.GetPerson())
}
