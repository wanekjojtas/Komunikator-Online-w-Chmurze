package user

import (
	"log"
	"net/http"
	"regexp"
	"server/util"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		Service: s,
	}
}

func isValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var user CreateUserReq
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Error binding CreateUser request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidEmail(user.Email) {
		log.Printf("Invalid email format: %s", user.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	res, err := h.Service.CreateUser(c.Request.Context(), &user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		if err.Error() == "Email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		if err.Error() == "Username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User created successfully: ID=%s, Username=%s", res.ID, res.Username)
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Login(c *gin.Context) {
	var user LoginUserReq
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Error binding Login request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if !isValidEmail(user.Email) {
		log.Printf("Invalid email format: %s", user.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	u, err := h.Service.Login(c.Request.Context(), &user)
	if err != nil {
		log.Printf("Error during login for email %s: %v", user.Email, err)
		if err.Error() == "Invalid password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		if err.Error() == "User not found" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Printf("Login successful: ID=%s, Username=%s", u.ID, u.Username)

	// Generate the JWT token
	token, err := util.GenerateAccessToken(u.ID, u.Username)
	if err != nil {
		log.Printf("Error generating token for user %s: %v", u.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"id":       u.ID,
		"username": u.Username,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "", "", false, true)
	log.Printf("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *Handler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		log.Printf("Missing query parameter for SearchUsers")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
		return
	}

	users, err := h.Service.SearchUsers(c.Request.Context(), query)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	log.Printf("SearchUsers successful for query: %s", query)
	c.JSON(http.StatusOK, users)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("Authorization")
	if refreshToken == "" {
		log.Printf("No refresh token provided")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No refresh token provided"})
		return
	}

	// Remove the "Bearer " prefix if present
	if len(refreshToken) > 7 && refreshToken[:7] == "Bearer " {
		refreshToken = refreshToken[7:]
	}

	log.Printf("Refreshing token for token: %s", refreshToken)

	// Validate the refresh token and generate a new access token
	newAccessToken, err := util.RefreshToken(refreshToken)
	if err != nil {
		log.Printf("Error refreshing token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	log.Printf("Token refreshed successfully")
	c.JSON(http.StatusOK, gin.H{"accessToken": newAccessToken})
}
