package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/amukherj/envoygrpc/names"
)

type NamesService struct{}

func (m NamesService) Get(ctx context.Context,
	in *names.NameRequest) (*names.NameResponse, error) {

	log.Printf(`[names] Names request received: \n
	From: %s
	Sent-at: %d\n`,
		in.GetServerName(), in.GetUtcTime())

	hostname, _ := os.Hostname()
	person := "Einstein"
	quality := "Brilliant"
	now := time.Now().Unix()
	out := &names.NameResponse{
		ServerName: &hostname,
		UtcTime:    &now,
		Person:     &person,
		Quality:    &quality,
	}

	return out, nil
}

func main() {
	port := ":50505"
	if len(os.Args) > 1 {
		port = os.Args[1]
		log.Println("len(os.Args) > 1: ", port)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("[names] Failed to start listening on %s: %v", port, err)
	}

	svrName, err := os.Hostname()
	svr := grpc.NewServer()
	var svc NamesService
	names.RegisterGreatNamesServiceServer(svr, svc)
	log.Printf("[names] Starting GRPC server on %s port %s ...\n", svrName, port)
	if err := svr.Serve(lis); err != nil {
		log.Fatalf("[names] Failed to start GRPC service: %v", err)
	}
}
