package internal

type Room struct {
	id    string
	peers map[string]*Client
	// hostID — id того, кто создал комнату (зашёл первым). Выставляется
	// один раз при создании Room и не меняется, даже если хост потом выйдет.
	hostID string
}

type inbound struct {
	client *Client
	msg    Envelope
}

// maxRoomPeers — жёсткий лимит на комнату. Пока держим только 1-на-1
// (mesh на N участников — отдельная, более крупная задача на будущее).
const maxRoomPeers = 2

type Hub struct {
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	inbound    chan inbound
}

func NewHub() *Hub {
	hub := &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		inbound:    make(chan inbound),
	}

	return hub
}

func (h *Hub) safeSend(peer *Client, msg Envelope) {
	select {
	case peer.send <- msg:
	default:
		h.leave(peer)
	}
}

func (h *Hub) join(c *Client) {
	roomId := c.room

	room := h.rooms[roomId]

	if room == nil {
		h.rooms[roomId] = &Room{
			id:     roomId,
			peers:  make(map[string]*Client),
			hostID: c.id,
		}
		room = h.rooms[roomId]
	}

	if len(room.peers) >= maxRoomPeers {
		h.safeSend(c, Envelope{Type: "room-full"})
		c.conn.Close()
		return
	}

	existing := make([]Peer, 0, len(room.peers))
	for id, peer := range room.peers {
		existing = append(existing, Peer{ID: id, Name: peer.name, IsHost: id == room.hostID})
	}

	room.peers[c.id] = c

	h.safeSend(c, Envelope{Type: "joined", From: c.id, Peers: existing, IsHost: c.id == room.hostID})

	for _, peer := range room.peers {
		if c.id != peer.id {
			message := Envelope{Type: "peer-joined", From: c.id, Name: c.name}
			h.safeSend(peer, message)
		}
	}
}

func (h *Hub) leave(c *Client) {
	roomId := c.room

	room := h.rooms[roomId]

	if room == nil {
		return
	}

	delete(room.peers, c.id)

	if c.active {
		close(c.send)
		c.active = false
	}

	if len(room.peers) == 0 {
		delete(h.rooms, roomId)
	}

	for _, peer := range room.peers {
		message := Envelope{Type: "peer-left", From: c.id}
		h.safeSend(peer, message)
	}
}

func (h *Hub) route(m inbound) {
	room := h.rooms[m.client.room]
	if room == nil {
		return
	}

	if dst, ok := room.peers[m.msg.To]; ok {
		h.safeSend(dst, m.msg)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.join(c)
		case c := <-h.unregister:
			h.leave(c)
		case m := <-h.inbound:
			h.route(m)
		}
	}
}
