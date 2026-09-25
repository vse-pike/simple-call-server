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
	// Name — отображаемое имя участника. Сервер его не интерпретирует,
	// только ретранслирует: клиент передаёт своё имя при подключении
	// (?name= в URL), сервер кладёт его в Client и рассылает в peer-joined.
	Name string `json:"name,omitempty"`
	// PeerNames — имена уже существующих в комнате участников (id -> name),
	// идёт вместе с Peers в "joined": Peers один даёт только id, по ним
	// новый участник не узнал бы, как их зовут.
	PeerNames map[string]string `json:"peerNames,omitempty"`
}
