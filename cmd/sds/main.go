package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	sds "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// SecretTypeURL defines the type URL for Envoy secret proto.
	SecretTypeURL = "type.googleapis.com/envoy.api.v2.auth.Secret"

	// SecretName defines the type of the secrets to fetch from the SDS server.
	SecretName = "server_cert"
)

type SDSImpl struct {
	certFile string
	keyFile  string
}

func NewSDS(certFile, keyFile string) *SDSImpl {
	return &SDSImpl{
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func (s *SDSImpl) GetTLSCertificate() (*auth.TlsCertificate, error) {
	certChain, err := ioutil.ReadFile(s.certFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s: %v", s.certFile, err)
	}

	keyContent, err := ioutil.ReadFile(s.keyFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s: %v", s.certFile, err)
	}

	tlsSecret := &auth.TlsCertificate{
		CertificateChain: &core.DataSource{
			Specifier: &core.DataSource_InlineBytes{certChain},
		},
		PrivateKey: &core.DataSource{
			Specifier: &core.DataSource_InlineBytes{keyContent},
		},
	}

	return tlsSecret, nil
}

func (s *SDSImpl) makeSecret() (*api.DiscoveryResponse, error) {
	tlsCertificate, err := s.GetTLSCertificate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read TLS certificate (%v)", err)
	}

	resources := make([]*any.Any, 1)
	secret := &auth.Secret{
		Name: SecretName,
		Type: &auth.Secret_TlsCertificate{
			TlsCertificate: tlsCertificate,
		},
	}
	data, err := proto.Marshal(secret)
	if err != nil {
		errMessage := fmt.Sprintf("Generates invalid secret (%v)", err)
		log.Println(errMessage)
		return nil, status.Errorf(codes.Internal, errMessage)
	}
	resources[0] = &any.Any{
		TypeUrl: SecretTypeURL,
		Value:   data,
	}

	response := &api.DiscoveryResponse{
		Resources:   resources,
		TypeUrl:     SecretTypeURL,
		VersionInfo: "v1",
	}

	return response, nil
}

func (s *SDSImpl) FetchSecrets(ctx context.Context, request *api.DiscoveryRequest) (*api.DiscoveryResponse, error) {
	return s.makeSecret()
}

func (s *SDSImpl) StreamSecrets(stream sds.SecretDiscoveryService_StreamSecretsServer) error {
	for {
		secret, _ := s.makeSecret()
		if err := stream.Send(secret); err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	}
	return nil
}

func (s *SDSImpl) DeltaSecrets(stream sds.SecretDiscoveryService_DeltaSecretsServer) error {
	err := "DeltaSecrets not implemented."
	log.Println(err)
	return status.Errorf(codes.Unimplemented, err)
}

func handleSignals(chnl <-chan os.Signal, svr *grpc.Server, udsPath string) {
	for sig := range chnl {
		go func(sig os.Signal) {
			if sig == syscall.SIGTERM || sig == syscall.SIGKILL {
				if _, err := os.Stat(udsPath); err == nil {
					// More likely that udsPath exists.
					// Likely, not certain because another
					// handler may have removed the file in
					// the meantime. But likelihood of existing
					// goes up with this check.
					//
					// We ignore errors
					svr.GracefulStop()
					os.Remove(udsPath)
				}
			}
		}(sig)
	}
}

func main() {
	baseDir := "/home/amukher1/devel/go/src/github.com/amukherj/envoygrpc"
	udsPath := baseDir + "/uds"
	_, err := os.Stat(udsPath)
	if err == nil {
		err = fmt.Errorf("The path %s already exists\n", udsPath)
		log.Fatalf("%v", err)
	}

	chsig := make(chan os.Signal, 1)

	listener, err := net.Listen("unix", udsPath)
	if err != nil {
		err = fmt.Errorf("Failed to listen on the UDS %s: %v", udsPath, err)
	}

	udsServer := grpc.NewServer()
	signal.Notify(chsig)
	go handleSignals(chsig, udsServer, udsPath)

	sds.RegisterSecretDiscoveryServiceServer(udsServer, NewSDS(
		baseDir+"/config/certs/envoy.crt",
		baseDir+"/config/certs/envoy.key"))
	if err = udsServer.Serve(listener); err != nil {
		log.Fatalf("Failed to launch SDS on UDS: %v", err)
	}
}
