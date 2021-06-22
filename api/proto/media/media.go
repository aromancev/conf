package media

//go:generate protoc media.proto --go_opt=Mmedia.proto=github.com/aromancev/confa/proto/media --proto_path=. --go_opt=paths=source_relative --go_out=. --twirp_out=.
