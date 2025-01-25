package ws

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

type Chat struct {
	ID       string             `json:"id"`
	Name     string             `json:"name,omitempty"`
	Members  map[string]*Client `json:"members"`
	Messages []*Message         `json:"messages"`
}

type Hub struct {
	mu         sync.RWMutex        // Protects access to Chats
	Chats      map[string]*Chat    // Active chats in the hub
	Register   chan *Client        // Channel for registering new clients
	Unregister chan *Client        // Channel for unregistering clients
	Broadcast  chan *Message       // Channel for broadcasting messages
	SyncChat   chan string         // Channel for synchronizing chats
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
		h.mu.Lock()
		h.Chats[chatID] = &Chat{
			ID:      chatID,
			Name:    chatName,
			Members: make(map[string]*Client),
		}
		h.mu.Unlock()
	}
	return nil
}

func NewHub() *Hub {
	return &Hub{
		Chats:      make(map[string]*Chat),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
		SyncChat:   make(chan string),
	}
}

func (h *Hub) Run(db *sql.DB) {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if chat, exists := h.Chats[client.RoomID]; exists {
				chat.Members[client.ID] = client
				log.Printf("Client %s joined chat %s", client.ID, client.RoomID)
			} else {
				log.Printf("Chat %s not found for client %s", client.RoomID, client.ID)
			}
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if chat, exists := h.Chats[client.RoomID]; exists {
				if _, exists := chat.Members[client.ID]; exists {
					delete(chat.Members, client.ID)
					close(client.Message)
					log.Printf("Client %s left chat %s", client.ID, client.RoomID)
				}
				if len(chat.Members) == 0 {
					log.Printf("No members left in chat %s. Chat can be archived or removed.", chat.ID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			if chat, exists := h.Chats[msg.RoomID]; exists {
				for _, client := range chat.Members {
					select {
					case client.Message <- msg:
					default:
						// Handle blocked message channels (e.g., disconnected clients)
						close(client.Message)
						h.mu.Lock()
						delete(chat.Members, client.ID)
						h.mu.Unlock()
						log.Printf("Client %s disconnected from chat %s", client.ID, chat.ID)
					}
				}
			} else {
				log.Printf("Chat %s not found for message broadcast", msg.RoomID)
			}
			h.mu.RUnlock()

		case chatID := <-h.SyncChat:
			h.mu.Lock()
			if _, exists := h.Chats[chatID]; !exists {
				var chatName string
				err := db.QueryRow("SELECT name FROM chats WHERE id = $1", chatID).Scan(&chatName)
				if err != nil {
					log.Printf("Failed to synchronize chat %s: %v", chatID, err)
				} else {
					h.Chats[chatID] = &Chat{
						ID:      chatID,
						Name:    chatName,
						Members: make(map[string]*Client),
					}
					log.Printf("Chat %s synchronized with hub", chatID)
				}
			}
			h.mu.Unlock()
		}
	}
}
