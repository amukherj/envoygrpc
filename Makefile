ALL=messages/messages.pb.go bin/server bin/client

all: prereq $(ALL)

.PHONY: prereq
prereq:
	if [ -z "${GOPATH}" ]; then \
	  echo "GOPATH is not set"; \
	  exit 1; \
	fi
	which protoc >/dev/null
	if [ $$? -ne 0 ]; then \
	  curl -OL https://github.com/google/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip; \
	  unzip protoc-3.6.1-linux-x86_64.zip -d /tmp/protoc3; \
	  sudo mv /tmp/protoc3/bin/* /usr/bin/; \
	  sudo mv /tmp/protoc3/include/* /usr/include; \
	fi
	go get google.golang.org/grpc

messages/messages.pb.go: protos/messages/messages.proto
	mkdir -p messages
	protoc -I $$(dirname $<) $$(basename $<) --go_out=plugins=grpc:$$(dirname $@)

bin/server: cmd/server/main.go messages/messages.pb.go
	go build -o bin/server ./$$(dirname $<)

bin/client: cmd/client/main.go messages/messages.pb.go
	go build -o bin/client ./$$(dirname $<)

clean:
	rm -rf $(ALL)
