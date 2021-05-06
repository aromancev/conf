package sfu

import (
	"github.com/pion/webrtc/v3"
)

//go:generate protoc --proto_path=. --go_opt=Msfu.proto=github.com/aromancev/confa/proto/sfu --go-grpc_opt=Msfu.proto=github.com/aromancev/confa/proto/sfu sfu.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=. --go_out=.

func CandidateInitFromRTC(init webrtc.ICECandidateInit) *CandidateInit {
	candidate := &CandidateInit{
		Candidate: init.Candidate,
	}
	if init.SDPMid != nil {
		candidate.SdpmIdSet = true
		candidate.SdpmId = *init.SDPMid
	}
	if init.SDPMLineIndex != nil {
		candidate.SdpmLineIndexSet = true
		candidate.SdpmLineIndex = int32(*init.SDPMLineIndex)
	}
	if init.UsernameFragment != nil {
		candidate.UsernameFragmentSet = true
		candidate.UsernameFragment = *init.UsernameFragment
	}
	return candidate
}

func CandidateInitToRTC(init *CandidateInit) webrtc.ICECandidateInit {
	candidate := webrtc.ICECandidateInit{
		Candidate:        init.Candidate,
		SDPMid:           nil,
		SDPMLineIndex:    nil,
		UsernameFragment: nil,
	}
	if init.SdpmIdSet {
		candidate.SDPMid = &init.SdpmId
	}
	if init.SdpmLineIndexSet {
		i := uint16(init.SdpmLineIndex)
		candidate.SDPMLineIndex = &i
	}
	if init.UsernameFragmentSet {
		candidate.UsernameFragment = &init.UsernameFragment
	}
	return candidate
}

func SessionDescriptionFromRTC(desc webrtc.SessionDescription) *SessionDescription {
	return &SessionDescription{
		Type: int32(desc.Type),
		Sdp:  desc.SDP,
	}
}

func SessionDescriptionToRTC(desc *SessionDescription) webrtc.SessionDescription {
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(desc.Type),
		SDP:  desc.Sdp,
	}
}
