package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/sourcegraph/jsonrpc2"
	"google.golang.org/grpc"

	psfu "github.com/aromancev/confa/proto/sfu"
)

func NewUpgrader(readBuf, writeBuf int) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  readBuf,
		WriteBufferSize: writeBuf,
	}
}

type Signal struct {
	sfuAddress string
	conn       *grpc.ClientConn
	stream     psfu.SFU_SignalClient

	lock sync.Mutex
}

func NewSignal(ctx context.Context, sfuAddress string) (*Signal, error) {
	conn, err := grpc.Dial(sfuAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to open GRPC conn: %w", err)
	}
	client := psfu.NewSFUClient(conn)
	stream, err := client.Signal(ctx)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open GRPC stream: %w", err)
	}

	return &Signal{
		sfuAddress: sfuAddress,
		conn:       conn,
		stream:     stream,
	}, nil
}

func (s *Signal) Serve(ctx context.Context, ws *jsonrpc2.Conn) error {
	for {
		reply, err := s.stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch payload := reply.Payload.(type) {
		case *psfu.SignalReply_Join:
			var body webrtc.SessionDescription
			err = json.Unmarshal(payload.Join, &body)
			if err != nil {
				panic(err)
			}
			if err := ws.Reply(ctx, jsonID(reply.Id), body); err != nil {
				panic(err)
			}

		case *psfu.SignalReply_Offer:
			var body webrtc.SessionDescription
			err = json.Unmarshal(payload.Offer, &body)
			if err != nil {
				panic(err)
			}
			if err := ws.Reply(ctx, jsonID(reply.Id), body); err != nil {
				panic(err)
			}

		case *psfu.SignalReply_Description:
			var body webrtc.SessionDescription
			err = json.Unmarshal(payload.Description, &body)
			if err != nil {
				panic(err)
			}
			if err := ws.Notify(ctx, methodOffer, body); err != nil {
				panic(err)
			}

		case *psfu.SignalReply_Trickle:
			var body trickle
			err = json.Unmarshal(payload.Trickle, &body)
			if err != nil {
				panic(err)
			}
			err := ws.Notify(ctx, methodTrickle, body)
			if err != nil {
				panic(err)
			}
		}
	}

	return s.stream.CloseSend()
}

func (s *Signal) Handle(_ context.Context, _ *jsonrpc2.Conn, request *jsonrpc2.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	switch request.Method {
	case methodJoin:
		err := s.stream.Send(&psfu.SignalRequest{
			Id: rpcID(request.ID),
			Payload: &psfu.SignalRequest_Join{
				Join: *request.Params,
			},
		})
		if err != nil {
			panic(err)
		}

	case methodOffer:
		err := s.stream.Send(&psfu.SignalRequest{
			Id: rpcID(request.ID),
			Payload: &psfu.SignalRequest_Offer{
				Offer: *request.Params,
			},
		})
		if err != nil {
			panic(err)
		}

	case methodAnswer:
		err := s.stream.Send(&psfu.SignalRequest{
			Payload: &psfu.SignalRequest_Answer{
				Answer: *request.Params,
			},
		})
		if err != nil {
			panic(err)
		}

	case methodTrickle:
		err := s.stream.Send(&psfu.SignalRequest{
			Payload: &psfu.SignalRequest_Trickle{
				Trickle: *request.Params,
			},
		})
		if err != nil {
			panic(err)
		}
	}
}

func (s *Signal) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_ = s.stream.CloseSend()
	return s.conn.Close()
}

const (
	methodJoin    = "join"
	methodOffer   = "offer"
	methodAnswer  = "answer"
	methodTrickle = "trickle"
)

// trickle message sent when renegotiating the peer connection.
type trickle struct {
	Target    int                     `json:"target"`
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

func rpcID(id jsonrpc2.ID) string {
	if id.IsString {
		return id.Str
	}
	return fmt.Sprintf("%d", id.Num)
}

func jsonID(id string) jsonrpc2.ID {
	return jsonrpc2.ID{Str: id, IsString: true}
}
