# Using Envoy as a reverse proxy for GRPC services

This repo demonstrates how to configure Envoy for routing to gRPC services.

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

#### TLS
If you want to test TLS support, start Envoy thus:

    ./bin/envoy -c config/envoy/tls/envoy.yaml

Run the client:

    ./bin/client 0.0.0.0:9943 "Your message here"

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
