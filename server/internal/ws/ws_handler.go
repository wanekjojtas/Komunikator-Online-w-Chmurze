package ws

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"server/util"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
)

type Handler struct {
	hub *Hub
	db  *sql.DB
}

func NewHandler(h *Hub, db *sql.DB) *Handler {
	return &Handler{hub: h, db: db}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ValidateToken validates the JWT token.
func (h *Handler) ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	claims, err := util.ValidateToken(token,false)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	log.Printf("Token validated successfully. Claims: %+v", claims)
	c.JSON(http.StatusOK, gin.H{"status": "Token is valid"})
}

// StartChat creates a new chat if it doesn't already exist.
func (h *Handler) StartChat(c *gin.Context) {
    var req struct {
        Members []string `json:"members"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
        return
    }

    if len(req.Members) < 1 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "At least one member is required to start a chat"})
        return
    }

    // Ensure unique and sorted members
    req.Members = uniqueMembers(req.Members)
    sort.Strings(req.Members)

    // Get the requesting user's ID
    requestingUserID := c.GetString("userID")
    if requestingUserID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Check if a one-on-one or group chat already exists
    query := `
        SELECT c.id
        FROM chats c
        INNER JOIN chat_members cm ON c.id = cm.chat_id
        WHERE c.id IN (
            SELECT chat_id
            FROM chat_members
            WHERE user_id = ANY($1::uuid[])
            GROUP BY chat_id
            HAVING COUNT(*) = $2
        )
    `
    var existingChatID string
    err := h.db.QueryRow(query, pq.Array(req.Members), len(req.Members)).Scan(&existingChatID)
    if err == nil {
        // Chat exists, ensure the requesting user is a member
        log.Printf("Chat already exists: ChatID=%s", existingChatID)

        var isMember bool
        err := h.db.QueryRow(`
            SELECT EXISTS (
                SELECT 1 
                FROM chat_members 
                WHERE chat_id = $1 AND user_id = $2
            )
        `, existingChatID, requestingUserID).Scan(&isMember)
        if err != nil {
            log.Printf("Error checking user membership: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        if !isMember {
            // Add the requesting user to the existing chat
            _, err = h.db.Exec(`
                INSERT INTO chat_members (chat_id, user_id) 
                VALUES ($1, $2)
            `, existingChatID, requestingUserID)
            if err != nil {
                log.Printf("Error adding user to existing chat: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join existing chat"})
                return
            }
            log.Printf("User %s added to existing chat %s", requestingUserID, existingChatID)
        }

        // Fetch the chat name (other user's username for one-on-one chat)
        var chatName string
        if len(req.Members) == 2 {
            var otherUserID string
            for _, member := range req.Members {
                if member != requestingUserID {
                    otherUserID = member
                    break
                }
            }

            var otherUsername string
            err = h.db.QueryRow("SELECT username FROM users WHERE id = $1", otherUserID).Scan(&otherUsername)
            if err != nil {
                log.Printf("Error fetching username for one-on-one chat: %v", err)
                chatName = "Unknown"
            } else {
                chatName = otherUsername
            }
        } else {
            chatName = "Group Chat"
        }

        c.JSON(http.StatusOK, gin.H{"chatID": existingChatID, "name": chatName})
        return
    }

    // If no existing chat, create a new one
    chatID := uuid.New().String()
    log.Printf("Creating new chat: ChatID=%s", chatID)

    var chatName string
    if len(req.Members) == 2 {
        // One-on-one chat: Determine the other user
        var otherUserID string
        for _, member := range req.Members {
            if member != requestingUserID {
                otherUserID = member
                break
            }
        }

        var otherUsername string
        err = h.db.QueryRow("SELECT username FROM users WHERE id = $1", otherUserID).Scan(&otherUsername)
        if err != nil {
            log.Printf("Error fetching username for new one-on-one chat: %v", err)
            chatName = "Unknown"
        } else {
            chatName = otherUsername
        }
    } else if len(req.Members) > 2 {
        chatName = "Group Chat"
    } else {
        chatName = "Chat with Yourself"
    }

    // Insert the new chat
    _, err = h.db.Exec("INSERT INTO chats (id, name) VALUES ($1, $2)", chatID, chatName)
    if err != nil {
        log.Printf("Error creating new chat: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
        return
    }

    // Add members to the new chat
    tx, err := h.db.Begin()
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to panic: %v", r)
        }
    }()

    for _, memberID := range req.Members {
        _, err := tx.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)", chatID, memberID)
        if err != nil {
            tx.Rollback()
            log.Printf("Error adding member: ChatID=%s, MemberID=%s, Error=%v", chatID, memberID, err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add members"})
            return
        }
    }

    if err := tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"chatID": chatID, "name": chatName, "members": req.Members})
}

// Helper function to remove duplicate IDs
func uniqueMembers(members []string) []string {
    unique := make(map[string]bool)
    var result []string
    for _, member := range members {
        if !unique[member] {
            unique[member] = true
            result = append(result, member)
        }
    }
    return result
}

// JoinChat establishes a WebSocket connection to a chat room.
func (h *Handler) JoinChat(c *gin.Context) {
    chatID := c.Param("chatID")
    userID := c.Query("userID")
    username := c.Query("username")
    token := c.Query("token")

    log.Printf("WebSocket JoinChat Request: chatID=%s, userID=%s, username=%s", chatID, userID, username)

    // Validate parameters and token
    if chatID == "" || userID == "" || username == "" || token == "" {
        log.Println("Missing required parameters")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
        return
    }

    claims, err := util.ValidateToken(token,false)
    if err != nil || claims.ID != userID || claims.Username != username {
        log.Printf("Invalid token or token mismatch for user: %s", username)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Validate that the chat exists in the `chats` table
    var chatExists bool
    err = h.db.QueryRow("SELECT EXISTS (SELECT 1 FROM chats WHERE id = $1)", chatID).Scan(&chatExists)
    if err != nil {
        log.Printf("Error checking chat existence for ChatID=%s: %v", chatID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    if !chatExists {
        log.Printf("Chat not found in database: ChatID=%s", chatID)
        c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
        return
    }

    // Synchronize chat with hub if missing
    if _, exists := h.hub.Chats[chatID]; !exists {
        log.Printf("Chat %s not found in hub, synchronizing with database", chatID)

        var chatName string
        err := h.db.QueryRow("SELECT name FROM chats WHERE id = $1", chatID).Scan(&chatName)
        if err != nil {
            log.Printf("Error fetching chat name for ChatID=%s: %v", chatID, err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        h.hub.Chats[chatID] = &Chat{
            ID:      chatID,
            Name:    chatName,
            Members: make(map[string]*Client),
        }
        log.Printf("Chat %s synchronized with hub", chatID)
    }

    // Validate user membership in the chat
    var isMember bool
    err = h.db.QueryRow(`
        SELECT EXISTS (
            SELECT 1 
            FROM chat_members 
            WHERE chat_id = $1 AND user_id = $2
        )
    `, chatID, userID).Scan(&isMember)
    if err != nil {
        log.Printf("Error validating user membership: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    if !isMember {
        log.Printf("User %s is not a member of chat %s", userID, chatID)
        c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("User %s is not a member of chat %s", userID, chatID)})
        return
    }

    // WebSocket connection upgrade
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
        return
    }

    client := &Client{
        Conn:     conn,
        Message:  make(chan *Message, 10),
        ID:       userID,
        RoomID:   chatID,
        Username: username,
        DB:       h.db,
    }

    // Add client to the chat room
    h.hub.Chats[chatID].Members[userID] = client
    log.Printf("User %s joined chat %s", username, chatID)

    defer func() {
        delete(h.hub.Chats[chatID].Members, userID)
        client.Conn.Close()
    }()

    // Start WebSocket read/write handling
    go client.writeMessage()
    client.readMessage(h.hub)
}

func (h *Handler) GetUserChats(c *gin.Context) {
    userID := c.GetString("userID")
    if userID == "" {
        log.Println("Missing userID in context")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    query := `
        SELECT 
            c.id,
            CASE
                WHEN COUNT(DISTINCT u.id) = 1 THEN 'Chat with Yourself' -- Self-chat
                WHEN COUNT(DISTINCT u.id) = 2 THEN (
                    SELECT username 
                    FROM users u2 
                    WHERE u2.id = (
                        SELECT user_id 
                        FROM chat_members 
                        WHERE chat_id = c.id 
                        AND user_id != $1
                        LIMIT 1
                    )
                ) -- One-on-one chat
                WHEN COUNT(DISTINCT u.id) > 2 THEN 'Group Chat' -- Group chat
                ELSE 'Unknown Chat'
            END AS name
        FROM chats c
        INNER JOIN chat_members cm ON c.id = cm.chat_id
        INNER JOIN users u ON cm.user_id = u.id
        WHERE c.id IN (
            SELECT chat_id FROM chat_members WHERE user_id = $1
        )
        GROUP BY c.id
        ORDER BY c.id;
    `

    rows, err := h.db.Query(query, userID)
    if err != nil {
        log.Printf("Error fetching user chats for userID=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user chats"})
        return
    }
    defer rows.Close()

    var chats []Chat
    for rows.Next() {
        var chat Chat
        if err := rows.Scan(&chat.ID, &chat.Name); err != nil {
            log.Printf("Error scanning user chats for userID=%s: %v", userID, err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse chat data"})
            return
        }
        chats = append(chats, chat)
    }

    c.JSON(http.StatusOK, chats)
}

func (h *Handler) GetChatDetails(c *gin.Context) {
    chatID := c.Param("chatID")
    if chatID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chatID"})
        return
    }

    userID := c.GetString("userID")
    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    var chat struct {
        ID      string   `json:"id"`
        Name    string   `json:"name"`
        Members []string `json:"members"`
    }

    query := `
        SELECT c.id, c.name, ARRAY_AGG(cm.user_id) AS members
        FROM chats c
        LEFT JOIN chat_members cm ON c.id = cm.chat_id
        WHERE c.id = $1
        GROUP BY c.id
    `
    err := h.db.QueryRow(query, chatID).Scan(&chat.ID, &chat.Name, pq.Array(&chat.Members))
    if err != nil {
        log.Printf("Error fetching chat details for chatID: %s, Error: %v", chatID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat details"})
        return
    }

    // Determine chat name dynamically for one-on-one chats
    if len(chat.Members) == 2 {
        var otherUserID string
        for _, member := range chat.Members {
            if member != userID {
                otherUserID = member
                break
            }
        }

        var otherUsername string
        err := h.db.QueryRow("SELECT username FROM users WHERE id = $1", otherUserID).Scan(&otherUsername)
        if err != nil {
            log.Printf("Error fetching username for chat: %v", err)
            chat.Name = "Chat"
        } else {
            chat.Name = otherUsername
        }
    }

    c.JSON(http.StatusOK, chat)
}

func (h *Handler) SendMessage(c *gin.Context) {
    var req struct {
        ChatID  string `json:"chatID"`
        Content string `json:"content"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        log.Printf("Invalid request payload: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
        return
    }

    userID := c.GetString("userID")
    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Validate that the `chatID` exists in the `chats` table
    var exists bool
    err := h.db.QueryRow("SELECT EXISTS (SELECT 1 FROM chats WHERE id = $1)", req.ChatID).Scan(&exists)
    if err != nil {
        log.Printf("Error checking chat existence: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    if !exists {
        log.Printf("Chat ID does not exist: %s", req.ChatID)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Chat does not exist"})
        return
    }

    // Validate user membership in the chat
    var isMember bool
    err = h.db.QueryRow(`
        SELECT EXISTS (
            SELECT 1 
            FROM chat_members 
            WHERE chat_id = $1 AND user_id = $2
        )`, req.ChatID, userID).Scan(&isMember)
    if err != nil {
        log.Printf("Error validating user membership: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }
    if !isMember {
        log.Printf("User %s is not a member of chat %s", userID, req.ChatID)
        c.JSON(http.StatusForbidden, gin.H{"error": "User not a member of this chat"})
        return
    }

    // Insert the message
    query := `INSERT INTO messages (chat_id, sender_id, content) VALUES ($1, $2, $3) RETURNING id, created_at`
    var message struct {
        ID        string    `json:"id"`
        CreatedAt time.Time `json:"created_at"`
    }

    err = h.db.QueryRow(query, req.ChatID, userID, req.Content).Scan(&message.ID, &message.CreatedAt)
    if err != nil {
        log.Printf("Failed to save message: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
        return
    }

    // Broadcast the message to WebSocket clients
    msg := &Message{
        ID:        message.ID,
        RoomID:    req.ChatID,
        SenderID:  userID,
        Username:  c.GetString("username"),
        Content:   req.Content,
        CreatedAt: message.CreatedAt,
    }

    h.hub.Broadcast <- msg

    c.JSON(http.StatusOK, gin.H{
        "id":         message.ID,
        "content":    req.Content,
        "created_at": message.CreatedAt,
        "sender_id":  userID,
    })
}


func (h *Handler) GetChatMessages(c *gin.Context) {
	chatID := c.Param("chatID")

	rows, err := h.db.Query(`
		SELECT 
			m.id, 
			m.sender_id, 
			u.username, 
			m.content, 
			m.created_at 
		FROM 
			messages m
		INNER JOIN 
			users u ON m.sender_id = u.id
		WHERE 
			m.chat_id = $1 
		ORDER BY 
			m.created_at ASC`, chatID)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Username, &msg.Content, &msg.CreatedAt); err != nil {
			log.Printf("Error parsing messages: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse messages"})
			return
		}
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}

func (h *Handler) GetAllUsers(c *gin.Context) {
    token := c.GetHeader("Authorization")
    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
        return
    }

    token = strings.TrimPrefix(token, "Bearer ")
    claims, err := util.ValidateToken(token,false)
    if err != nil {
        log.Printf("Token validation failed: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
        return
    }

    log.Printf("Token validated successfully. UserID: %s", claims.ID)

    rows, err := h.db.Query("SELECT id, username FROM users")
    if err != nil {
        log.Printf("Error fetching users: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
        return
    }
    defer rows.Close()

    var users []struct {
        ID       string `json:"id"`
        Username string `json:"username"`
    }

    for rows.Next() {
        var user struct {
            ID       string `json:"id"`
            Username string `json:"username"`
        }
        if err := rows.Scan(&user.ID, &user.Username); err != nil {
            log.Printf("Error scanning user: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse users"})
            return
        }
        users = append(users, user)
    }

    c.JSON(http.StatusOK, users)
}


