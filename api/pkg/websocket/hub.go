package websocket

type Hub struct {
	clients    map[string]map[*Client]bool // caseID -> clients
	broadcast  chan MessageEnvelope
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan MessageEnvelope),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.CaseID] == nil {
				h.clients[client.CaseID] = make(map[*Client]bool)
			}
			h.clients[client.CaseID][client] = true

		case client := <-h.unregister:
			if clients, ok := h.clients[client.CaseID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
				}
			}

		case message := <-h.broadcast:
			if clients, ok := h.clients[message.CaseID]; ok {
				for client := range clients {
					select {
					case client.Send <- message.Data:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
		}
	}
}
