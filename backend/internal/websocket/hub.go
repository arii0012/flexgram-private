package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID string
}

type Hub struct {
	clients    map[string]map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("[WS] user %s connected", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WS] user %s disconnected", client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, clients := range h.clients {
				for client := range clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToUser(userID string, msgType string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[userID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) SendToUsers(userIDs []string, msgType string, payload interface{}) {
	for _, uid := range userIDs {
		h.SendToUser(uid, msgType, payload)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if t, ok := msg["type"].(string); ok && t == "ping" {
				c.Send <- []byte(`{"type":"pong"}`)
			}
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
