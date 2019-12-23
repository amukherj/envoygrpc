ALL=messages/messages.pb.go bin/server bin/client

all: $(ALL)

messages/messages.pb.go: protos/messages/messages.proto
	mkdir -p messages
	protoc -I $$(dirname $<) $$(basename $<) --go_out=plugins=grpc:$$(dirname $@)

bin/server: cmd/server/main.go messages/messages.pb.go
	go build -o bin/server ./$$(dirname $<)

bin/client: cmd/client/main.go messages/messages.pb.go
	go build -o bin/client ./$$(dirname $<)

clean:
	rm -rf $(ALL)
