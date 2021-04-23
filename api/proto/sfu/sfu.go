package sfu

//go:generate protoc --proto_path=. --go_opt=Msfu.proto=github.com/aromancev/confa/proto/sfu --go-grpc_opt=Msfu.proto=github.com/aromancev/confa/proto/sfu sfu.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=. --go_out=.
