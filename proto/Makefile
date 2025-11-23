.PHONY: all user file clean install-tools

all: user file

install-tools:
	@echo "Installing protobuf tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

user:
	@echo "Generating user service protobuf..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		user/user.proto

file:
	@echo "Generating file service protobuf..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		file/file.proto

clean:
	@echo "Cleaning generated files..."
	rm -f user/*.pb.go
	rm -f file/*.pb.go
