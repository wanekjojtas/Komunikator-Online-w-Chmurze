package ws

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
	ID       string `json:"ID"`
	RoomID   string `json:"roomID"`
	Username string `json:"username"`
	DB       *sql.DB
}

type Message struct {
    ID        string    `json:"id"`        // Message ID
    RoomID    string    `json:"roomID"`    // Chat/Room ID
    SenderID  string    `json:"senderID"`  // Sender's user ID
    Username  string    `json:"username"`  // Sender's username
    Content   string    `json:"content"`   // Message content
    CreatedAt time.Time `json:"createdAt"` // Timestamp of the message
}


func (c *Client) writeMessage() {
	defer c.Conn.Close()

	for msg := range c.Message {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Error writing WebSocket message to client %s: %v", c.ID, err)
			return
		}
		log.Printf("Message sent to client %s: %+v", c.ID, msg)
	}
}


func (c *Client) readMessage(hub *Hub) {
    defer func() {
        // Ensure cleanup on disconnect
        log.Printf("Client %s disconnected from chat %s", c.ID, c.RoomID)
        hub.Unregister <- c
        c.Conn.Close()
    }()

    for {
        // Read the message from the WebSocket connection
        _, messageBytes, err := c.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("Unexpected WebSocket closure for client %s: %v", c.ID, err)
            } else {
                log.Printf("WebSocket closed normally for client %s: %v", c.ID, err)
            }
            break
        }

        // Create a new message object
        msg := &Message{
            ID:        uuid.New().String(),
            RoomID:    c.RoomID,
            SenderID:  c.ID,
            Username:  c.Username,
            Content:   string(messageBytes),
            CreatedAt: time.Now(),
        }

        // Save the message to the database
        _, dbErr := c.DB.Exec(
            "INSERT INTO messages (id, chat_id, sender_id, content, created_at) VALUES ($1, $2, $3, $4, $5)",
            msg.ID, msg.RoomID, msg.SenderID, msg.Content, msg.CreatedAt,
        )
        if dbErr != nil {
            log.Printf("Failed to save message from client %s to database: %v", c.ID, dbErr)
        } else {
            log.Printf("Message saved to database: %+v", msg)
        }

        // Broadcast the message to other clients in the chat room
        hub.Broadcast <- msg
    }
}


