package handler

import (
	"context"
	"encoding/json"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/jsonrpc2"
)

type JSONSignal struct {
	*sfu.Peer
}

func NewJSONSignal(p *sfu.Peer) *JSONSignal {
	return &JSONSignal{p}
}

// Handle incoming RPC call events like join, answer, offer and trickle.
func (p *JSONSignal) Handle(ctx context.Context, conn *jsonrpc2.Conn, request *jsonrpc2.Request) {
	switch request.Method {
	case methodJoin:
		var req join
		if err := json.Unmarshal(*request.Params, &req); err != nil {
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "invalid payload",
			})
			return
		}

		p.OnOffer = func(offer *webrtc.SessionDescription) {
			if err := conn.Notify(ctx, methodOffer, offer); err != nil {
				log.Ctx(ctx).Err(err).Msg("failed to send offer")
			}
		}
		p.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
			if err := conn.Notify(ctx, methodTrickle, trickle{
				Candidate: *candidate,
				Target:    target,
			}); err != nil {
				log.Ctx(ctx).Err(err).Msg("failed to send trickle")
			}
		}

		if err := p.Join(req.Sid); err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to join")
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "failed to join",
			})
			return
		}

		answer, err := p.Answer(req.Offer)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to send answer")
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "failed to answer",
			})
			return
		}

		_ = conn.Reply(ctx, request.ID, answer)

	case methodOffer:
		var req negotiation
		if err := json.Unmarshal(*request.Params, &req); err != nil {
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "invalid payload",
			})
			return
		}

		answer, err := p.Answer(req.Desc)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to send answer")
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "failed to answer",
			})
			return
		}
		_ = conn.Reply(ctx, request.ID, answer)

	case methodAnswer:
		var req negotiation
		if err := json.Unmarshal(*request.Params, &req); err != nil {
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "invalid payload",
			})
			return
		}

		if err := p.SetRemoteDescription(req.Desc); err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to set remote description")
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "failed to set remote description",
			})
		}

	case methodTrickle:
		var req trickle
		err := json.Unmarshal(*request.Params, &req)
		if err != nil {
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "invalid payload",
			})
			break
		}

		err = p.Trickle(req.Candidate, req.Target)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to trickle")
			_ = conn.ReplyWithError(ctx, request.ID, &jsonrpc2.Error{
				Message: "failed to trickle",
			})
		}
	}
}

const (
	methodJoin    = "join"
	methodOffer   = "offer"
	methodAnswer  = "answer"
	methodTrickle = "trickle"
)

// join message sent when initializing a peer connection
type join struct {
	Sid   string                    `json:"sid"`
	Offer webrtc.SessionDescription `json:"offer"`
}

// negotiation message sent when renegotiating the peer connection
type negotiation struct {
	Desc webrtc.SessionDescription `json:"desc"`
}

// trickle message sent when renegotiating the peer connection
type trickle struct {
	Target    int                     `json:"target"`
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}
