# Using Envoy as a reverse proxy for GRPC services

This repo demonstrates how to configure Envoy for routing to gRPC services.
This doesn't yet support TLS-secured gRPC endpoints, which would cause minor
changes to the configuration.

## Building
To build the binaries just do the following.

    make

You will also need to get hold of the envoy binary from somewhere. Consider
pulling the Envoy docker image, running, and `docker cp`-ing the envoy
binary from inside it. Copy this binary to the `bin/` subdirectory of the repo.

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

### Egress
Start two instances of the gRPC server on a remote server. Note the IP
address of the remote server.

Edit the file config/envoy/envoy.yaml in the repo, and replace the IP address
`192.168.87.3` with the IP of your remote server. Now start Envoy locally:

    ./bin/envoy -c config/envoy/envoy.yaml

Run the client:

    ./bin/client 0.0.0.0:9911 "Your message here" authority emery

In the above, `emery` is the identifier for your remote host(s). It can be any
name as long as you also update it in the config.

In the response printed on the console, check if the From field is correctly set
to the remote server's hostname.
