package ws

import (
	"database/sql"
	"fmt"
)

type Chat struct {
	ID       string             `json:"id"`
	Name     string             `json:"name,omitempty"`
	Members  map[string]*Client `json:"members"`
	Messages []*Message         `json:"messages"`
}

type Hub struct {
	Chats      map[string]*Chat
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
}

func LoadChatsIntoHub(h *Hub, db *sql.DB) error {
	rows, err := db.Query("SELECT id, name FROM chats")
	if err != nil {
		return fmt.Errorf("failed to load chats: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chatID, chatName string
		if err := rows.Scan(&chatID, &chatName); err != nil {
			return fmt.Errorf("failed to scan chat: %v", err)
		}
		h.Chats[chatID] = &Chat{
			ID:      chatID,
			Name:    chatName,
			Members: make(map[string]*Client),
		}
	}
	return nil
}

func NewHub() *Hub {
	return &Hub{
		Chats:      make(map[string]*Chat),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
	}
}

func (h *Hub) Run() {
	for msg := range h.Broadcast {
		if chat, exists := h.Chats[msg.RoomID]; exists {
			for _, client := range chat.Members {
				select {
				case client.Message <- msg:
				default:
					// Handle blocked message channels (e.g., disconnected clients)
					close(client.Message)
					delete(chat.Members, client.ID)
				}
			}
		}
	}
}
