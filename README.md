# Using Envoy as a reverse proxy for GRPC services

This repo demonstrates how to configure [Envoy](https://www.envoyproxy.io/) for
routing to gRPC services. The focus is to show basic constructs for enabling
routing to gRPC services, making it work with TLS / mTLS (todo), and making
certificates available via the Secrets Discovery Service.

The norm for most such repos is to use at the least Docker. I have deliberately
avoided any form of containers or other deployment shebang to keep the focus
on just Envoy and make it utterly easy to understand what's going on.

## Code
1. The grpc service and message definitions are under `messages`.
2. The grpc service implementation is under `cmd/server`.
3. The grpc client implementation is under `cmd/client`.
4. The Secrets Discovery Service implementation is under `cmd/sds` and is
deliberately kept simple.

Self-signed certificates are automatically generated as part of the build
process. Look at the Makefile to understand what's going on. This also means
that you should have OpenSSL installed on your dev box.

## Building
To build the binaries just do the following.

    make

It is expected that you will copy a pre-built Envoy binary from somewhere into
`./bin`.  Consider pulling the Envoy docker image, running, and `docker cp`-ing
the envoy binary from inside it. Copy this binary to the `bin/` subdirectory of
the repo.

## Running

### Ingress
Start two instances of the gRPC server locally:

    ./bin/server :50501
    ./bin/server :50503

Start Envoy:

    ./bin/envoy -c config/envoy/envoy.yaml

Run the client:

    ./bin/client 0.0.0.0:9911 "Your message here"

In the response printed on the console, check if the From field is correctly set
to the local host's hostname.

#### TLS
If you want to test TLS support, start Envoy thus:

    ./bin/envoy -c config/envoy/tls/envoy.yaml

Run the client:

    ./bin/client 0.0.0.0:9943 "Your message here"

#### TLS via Secrets Discovery Service (SDS)
You can serve TLS certs via the Secrets Discovery Service (SDS) instead of
statically. There is a simplistic SDS implementation in cmd/sds/main.go. To
test this, run the following commands in addition to starting the two
instances of ./bin/server on 50501 and 50503.

	./sds
    ./bin/envoy -c config/envoy/tls/envoy-sds.yaml

Run the client:

    ./bin/client 0.0.0.0:9943 "Your message here"

QQ: Why don't we use the envoyproxy/go-control-plane implementation of SDS?
Mainly because it doesn't support SDS connections via Unix domain sockets
and require that you set up mTLS between Envoy and the control plane process
running the SDS implementation. Feel free to fork this repo and try it.

### Egress
Start two instances of the gRPC server on a remote server. Note the IP
address of the remote server.

On the remote site, start an Envoy instance by running:

    ./bin/envoy -c config/envoy/envoy-emery.yaml

On the local server, edit config/envoy/envoy.yaml in the repo and replace
the IP address `192.168.87.*` with the IP of your remote server. Now start
Envoy locally:

    ./bin/envoy -c config/envoy/envoy.yaml

Run the client to connect to the egress port on the local Envoy:

    ./bin/client localhost:9912 "Your message here" authority emery

In the above, `emery` is the identifier for your remote host(s). It can be any
name as long as you also update it in the config.

In the response printed on the console, check if the From field is correctly set
to the remote server's hostname.

#### TLS
If you want to test TLS support, start Envoy on the remote server thus:

    make
    ./bin/envoy -c config/envoy/tls/envoy-emery.yaml

On the local server, start Envoy thus.

    ./bin/envoy -c config/envoy/tls/envoy.yaml

Run the client to connect to the egress port on the local Envoy:

    ./bin/client 0.0.0.0:9912 "Your message here" authority emery

In the above, `emery` is the identifier for your remote host(s). It can be any
name as long as you also update it in the config.

#### TLS via Secrets Discovery Service (SDS)
Left as an exercise. Easy to extend based on the earlier example.
