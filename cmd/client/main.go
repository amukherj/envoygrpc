package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
  "flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/amukherj/envoygrpc/messages"
	// "github.com/amukherj/envoygrpc/names"
)

var address string = "localhost:50501"

func GetTLSCreds(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
  if certFile == "" && keyFile == "" {
    return insecure.NewCredentials(), nil
  }

  cert, err := tls.LoadX509KeyPair(certFile, keyFile)
  if err != nil {
    return nil, fmt.Errorf("failed to load certificate: %w", err)
  }

  tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
  }

  if caFile != "" {
    ca, err := ioutil.ReadFile(caFile)
    if err != nil {
      return nil, fmt.Errorf("failed to read CA file: %w", err)
    }

    tlsConfig.RootCAs = x509.NewCertPool()
    if !tlsConfig.RootCAs.AppendCertsFromPEM(ca) {
      return nil, fmt.Errorf("failed to append certs from CA file")
    }
  } else {
    tlsConfig.InsecureSkipVerify = true
  }

  return credentials.NewTLS(tlsConfig), err
}

func main() {
  var keyFile, crtFile, caFile string
  flag.StringVar(&keyFile, "key", "", "Path to private key")
  flag.StringVar(&crtFile, "cert", "", "Path to certificate")
  flag.StringVar(&caFile, "cacert", "", "Path to CA file")
  flag.Parse();
	msg := "Go rules!"

  flagOffset := len(os.Args) - flag.NArg() -1

	if len(os.Args) <= flagOffset + 1 {
		log.Fatalf("usage: %s [-key <keyfile> -cert <certfile> [-cacert <ca_file>]] <server-IP:port> [[msg] header value]",
			os.Args[flag.NArg() + 0])
	}
	address = os.Args[flagOffset + 1]

	if len(os.Args) > flagOffset + 2 {
		msg = os.Args[flagOffset + 2]
	}

	var header, headerVal string
	if len(os.Args) > flagOffset + 4 {
		header = os.Args[flagOffset + 3]
		headerVal = os.Args[flagOffset + 4]
	}

	tlsCred, err := GetTLSCreds(crtFile, keyFile, caFile)
  if err != nil {
    log.Fatalf("Failed to load credentials: %v", err)
  }

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCred),
	}

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

	if len(header) > 0 {
    ctx = metadata.NewOutgoingContext(ctx,
      metadata.New(map[string]string{header: headerVal}))
	}

	resp, err := client.Hello(ctx, &payload)
	if err != nil {
		log.Fatalf("[client] RPC error: %v", err)
	}
	log.Printf(`[client] Response message:
	From: %s
	Sent-at: %d
	Msg: %s`, resp.GetServerName(), resp.GetUtcTime(), resp.GetMsg())

	// Call the GreatNamesService
	/* client1 := names.NewGreatNamesServiceClient(conn)
	now = time.Now().Unix()
	request := &names.NameRequest{
		ServerName: &serverName,
		UtcTime:    &now,
	}

	resp1, err := client1.Get(ctx, request)
	if err != nil {
		log.Fatalf("[client] RPC error: %v", err)
	}
	log.Printf(`[client] Response message:
	From: %s
	Sent-at: %d
	Response: %s %s`, resp1.GetServerName(), resp1.GetUtcTime(),
		resp1.GetQuality(), resp1.GetPerson()) */

	// Call the streaming WhatsUp service on EchoService
	stream, err := client.WhatsUp(ctx)

	msg_base := ""
	if len(os.Args) > 2 {
		msg_base = os.Args[2]
	}
	for i := 0; i < 10; i++ {
		now = time.Now().Unix()
		msg = fmt.Sprintf("%s at %d", msg_base, now)
		payload.UtcTime = &now
		payload.Msg = &msg
		if err = stream.Send(&payload); err != nil {
			log.Fatalf("Exiting: error while sending: %v", err)
		}
		if resp, err := stream.Recv(); err != nil {
			log.Fatalf("Exiting: error while receiving: %v", err)
		} else {
			log.Printf(`[client] Response message:
	From: %s
	Sent-at: %d
	Response: %s %s`, resp.GetServerName(), resp.GetUtcTime(), resp.GetMsg())
		}
    time.Sleep(100 * time.Millisecond)
	}
	stream.CloseSend()
}
