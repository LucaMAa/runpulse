package ws

import (
	"Chrono/config"
	"Chrono/middleware"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client rappresenta un singolo utente connesso via WebSocket
type Client struct {
	UserID uint
	Role   string // "start" | "end"
	conn   *websocket.Conn
	send   chan []byte
	room   *Room
	hub    *Hub
}

// ServeWS è l'handler Gin per la connessione WebSocket
// Query params richiesti: token=<JWT>, code=<codice stanza>, role=<start|end>
func ServeWS(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Autentica il token JWT
		tokenStr := c.Query("token")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token mancante"})
			return
		}

		claims := &middleware.Claims{}
		parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.Cfg.JWTSecret), nil
		})
		if err != nil || !parsed.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token non valido"})
			return
		}

		// 2. Valida parametri
		code := c.Query("code")
		role := c.Query("role")
		if code == "" || (role != "start" && role != "end") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "code e role (start|end) obbligatori"})
			return
		}

		// 3. Upgrade a WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("❌ Upgrade WebSocket fallito: %v", err)
			return
		}

		// 4. Trova o crea la stanza
		room := hub.GetOrCreateRoom(code)

		client := &Client{
			UserID: claims.UserID,
			Role:   role,
			conn:   conn,
			send:   make(chan []byte, 256),
			room:   room,
			hub:    hub,
		}

		room.Join(client)

		// Avvia le goroutine di lettura e scrittura
		go client.writePump()
		go client.readPump()
	}
}

// readPump legge i messaggi in arrivo e li fa broadcast nella stanza
func (c *Client) readPump() {
	defer func() {
		c.room.Leave(c)
		if c.room.Size() == 0 {
			c.hub.RemoveRoom(c.room.Code)
		}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ Errore WebSocket user %d: %v", c.UserID, err)
			}
			break
		}

		log.Printf("📩 [stanza %s] user %d (%s): %s", c.room.Code, c.UserID, c.Role, string(msg))
		c.room.Broadcast(c, msg)
	}
}

// writePump invia i messaggi in coda al client
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
