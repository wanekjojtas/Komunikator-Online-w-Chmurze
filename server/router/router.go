package router

import (
	"log"
	"os"
	"server/internal/middleware"
	"server/internal/user"
	"server/internal/ws"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func getAllowedOrigins() []string {
	// Allow different origins for development and production
	if gin.Mode() == gin.ReleaseMode {
		return []string{"https://golang-nextjs-chat-app-fe-f621f6c2c0af.herokuapp.com"}
	}
	return []string{"http://localhost:3000"}
}

func InitRouter(userHandler *user.Handler, wsHandler *ws.Handler) {
	r = gin.Default()

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Log all incoming requests
	r.Use(func(c *gin.Context) {
		log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
		if strings.HasPrefix(c.Request.URL.Path, "/ws/") {
			log.Printf("WebSocket Request Parameters: %v", c.Request.URL.Query())
		}
		c.Next()
	})

	// Public Routes
	r.POST("/signup", userHandler.CreateUser)
	r.POST("/login", userHandler.Login)
	r.POST("/auth/refresh-token", userHandler.RefreshToken) // Refresh Token Endpoint
	r.GET("/logout", userHandler.Logout)

	// Validate Token Route
	r.GET("/validate-token", middleware.AuthMiddleware(), wsHandler.ValidateToken)

	// User-related Authenticated Routes
	userRoutes := r.Group("/users", middleware.AuthMiddleware())
	{
		userRoutes.GET("/search", userHandler.SearchUsers)
		userRoutes.GET("/all", wsHandler.GetAllUsers)
	}

	// WebSocket-Related Authenticated Routes
	authRoutes := r.Group("/ws", middleware.AuthMiddleware())
	{
		authRoutes.POST("/startChat", wsHandler.StartChat)
		authRoutes.GET("/joinChat/:chatID", wsHandler.JoinChat)
		authRoutes.GET("/getUserChats", wsHandler.GetUserChats)
		authRoutes.GET("/getChatDetails/:chatID", wsHandler.GetChatDetails)
		authRoutes.POST("/sendMessage", wsHandler.SendMessage)
		authRoutes.GET("/getChatMessages/:chatID", wsHandler.GetChatMessages)
	}

	// OPTIONS Handler for Preflight Requests
	r.OPTIONS("/*path", func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if contains(getAllowedOrigins(), origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Connection", "Upgrade")
		c.Header("Upgrade", "websocket")
		c.AbortWithStatus(204)
	})
}

// Utility to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func Start(addr string) error {
	port := os.Getenv("PORT")
	if port != "" {
		addr = ":" + port
	}
	return r.Run(addr)
}
