package media

//go:generate protoc --proto_path=. --go_opt=Mmedia.proto=github.com/aromancev/confa/proto/media --go-grpc_opt=Mmedia.proto=github.com/aromancev/confa/proto/media media.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=. --go_out=.
