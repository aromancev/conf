package rtc

//go:generate protoc rtc.proto --go_opt=Mrtc.proto=github.com/aromancev/confa/proto/rtc --proto_path=. --go_opt=paths=source_relative --go_out=. --twirp_out=.
