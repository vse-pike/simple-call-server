package internal

import (
	"encoding/json"
)

// Peer — участник комнаты для envelope type="joined": id и имя вместе,
// одной сущностью, а не двумя параллельными структурами, которые пришлось
// бы держать в синхроне вручную.
type Peer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	IsHost bool   `json:"isHost,omitempty"`
}

type Envelope struct {
	Type   string          `json:"type"`
	From   string          `json:"from,omitempty"`
	To     string          `json:"to,omitempty"`
	RoomID string          `json:"roomId,omitempty"`
	SDP    json.RawMessage `json:"sdp,omitempty"`
	Cand   json.RawMessage `json:"candidate,omitempty"`
	Peers  []Peer          `json:"peers,omitempty"`
	// Name — отображаемое имя участника. Сервер его не интерпретирует,
	// только ретранслирует: клиент передаёт своё имя при подключении
	// (?name= в URL), сервер кладёт его в Client и рассылает в peer-joined.
	Name string `json:"name,omitempty"`
	// IsHost — хост ли From этого envelope. Актуально только для "joined"
	// (там From — это ты сам): хост комнаты — тот, кто её создал (зашёл
	// первым), см. Room.hostID в hub.go.
	IsHost bool `json:"isHost,omitempty"`
}
