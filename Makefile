.PHONY: compile

PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := $(shell which protoc)

UNAME := $(shell uname)

$(PROTOC):
ifeq ($(UNAME), Darwin)
	brew install protobuf
endif
ifeq ($(UNAME), Linux)
	sudo apt-get install protobuf-compiler
endif

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

build: $(PROTOC_GEN_GO) $(PROTOC)
	protoc -I protos/ -I /usr/local/include/ --go_out=plugins=grpc:protos --proto_path=protos/ protos/*.proto

compile: build

doc:
	godoc -http=:6060

test:
	go test registry/registry_test.go -v
