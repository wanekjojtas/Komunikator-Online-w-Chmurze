package user

import (
	"net/http"
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



func (h *Handler) CreateUser(c *gin.Context) {
    var user CreateUserReq
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    res, err := h.Service.CreateUser(c.Request.Context(), &user)
    if err != nil {
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

    c.JSON(http.StatusOK, res)
}


func (h *Handler) Login(c *gin.Context) {
    var user LoginUserReq
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
        return
    }

    u, err := h.Service.Login(c.Request.Context(), &user)
    if err != nil {
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

    // Generate the JWT token
    token, err := util.GenerateJWT(u.ID, u.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    // Construct and return the response
    c.JSON(http.StatusOK, gin.H{
        "token":    token,  // Include the JWT token in the response
        "id":       u.ID,
        "username": u.Username,
    })
}



func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *Handler) SearchUsers(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
        return
    }

    users, err := h.Service.SearchUsers(c.Request.Context(), query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
        return
    }

    c.JSON(http.StatusOK, users)
}


