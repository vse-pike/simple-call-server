package internal

import (
	"encoding/json"
)

type Envelope struct {
	Type   string          `json:"type"`
	From   string          `json:"from,omitempty"`
	To     string          `json:"to,omitempty"`
	RoomID string          `json:"roomId,omitempty"`
	SDP    json.RawMessage `json:"sdp,omitempty"`
	Cand   json.RawMessage `json:"candidate,omitempty"`
	Peers  []string        `json:"peers,omitempty"`
}
