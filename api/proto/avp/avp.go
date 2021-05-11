package avp

//go:generate protoc --proto_path=. --go_opt=Mavp.proto=github.com/aromancev/confa/proto/avp --go-grpc_opt=Mavp.proto=github.com/aromancev/confa/proto/avp avp.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=. --go_out=.
