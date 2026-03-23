package socket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"ai-symptom-checker/config"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == config.App.FrontendURL
	},
}

type Client struct {
	ConsultationID uuid.UUID
	Conn           *websocket.Conn
	Send           chan []byte
}

type MessageEvent struct {
	ConsultationID uuid.UUID   `json:"consultation_id"`
	Type           string      `json:"type"` // "new_message" | "typing"
	Data           interface{} `json:"data"`
}

type Hub struct {
	rooms      map[uuid.UUID]map[*Client]bool
	broadcast  chan MessageEvent
	register   chan *Client
	unregister chan *Client
	sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan MessageEvent),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			if h.rooms[client.ConsultationID] == nil {
				h.rooms[client.ConsultationID] = make(map[*Client]bool)
			}
			h.rooms[client.ConsultationID][client] = true
			h.Unlock()
			log.Printf("[SocketHub] Registered client for consultation: %s", client.ConsultationID)

		case client := <-h.unregister:
			h.Lock()
			if clients, ok := h.rooms[client.ConsultationID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.ConsultationID)
					}
				}
			}
			h.Unlock()
			log.Printf("[SocketHub] Unregistered client from consultation: %s", client.ConsultationID)

		case event := <-h.broadcast:
			h.RLock()
			if clients, ok := h.rooms[event.ConsultationID]; ok {
				payload, _ := json.Marshal(event)
				for client := range clients {
					select {
					case client.Send <- payload:
					default:
						// Handled in unregister
					}
				}
			}
			h.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToRoom(consultationID uuid.UUID, eventType string, data interface{}) {
	h.broadcast <- MessageEvent{
		ConsultationID: consultationID,
		Type:           eventType,
		Data:           data,
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, consultationID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[SocketHub] Upgrade error: %v", err)
		return
	}

	client := &Client{
		ConsultationID: consultationID,
		Conn:           conn,
		Send:           make(chan []byte, 256),
	}

	h.register <- client

	// Start pumps
	go client.writePump()
	go client.readPump(h)
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
