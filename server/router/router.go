package router

import (
	"log"
	"os"
	"server/internal/middleware"
	"server/internal/user"
	"server/internal/ws"
	"server/util"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func InitRouter(userHandler *user.Handler, wsHandler *ws.Handler) {
	r = gin.Default()
	r.SetTrustedProxies(nil) // Ensure headers are preserved in Heroku

	// Apply CORS middleware globally
	r.Use(cors.New(cors.Config{
		AllowOrigins:     util.GetAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Security Headers and Performance Improvements Middleware
	r.Use(func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")     // Prevent MIME sniffing
		c.Header("Cache-Control", "no-store")             // Disable caching sensitive data
		c.Header("Content-Type", "application/json; charset=utf-8") // Set charset
		c.Writer.Header().Del("X-Powered-By")             // Remove X-Powered-By header

		// Log incoming requests except OPTIONS
		if c.Request.Method != "OPTIONS" {
			log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
			if strings.HasPrefix(c.Request.URL.Path, "/ws/") {
				log.Printf("WebSocket Request Parameters: %v", c.Request.URL.Query())
			}
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
}

func Start(addr string) error {
	port := os.Getenv("PORT")
	if port != "" {
		addr = ":" + port
	}
	return r.Run(addr)
}
