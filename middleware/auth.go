package middleware

import (
	"Chrono/config"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Printf("🔑 Auth header: %q", authHeader)

		// Supporta anche token via query param per WebSocket
		// es: ws://host/ws?token=xxx&code=ABC123
		token := c.Query("token")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token mancante"})
			return
		}

		claims := &Claims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.Cfg.JWTSecret), nil
		})

		if err != nil || !parsed.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token non valido"})
			return
		}

		// Inietta i dati utente nel contesto Gin
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// Helper per estrarre userID dal contesto
func GetUserID(c *gin.Context) uint {
	id, _ := c.Get("userID")
	return id.(uint)
}
