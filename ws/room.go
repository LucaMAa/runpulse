package ws

import (
	"log"
	"sync"
)

// Room rappresenta una sessione di allenamento tra due client
type Room struct {
	Code    string
	clients map[*Client]bool
	mu      sync.RWMutex
}

func newRoom(code string) *Room {
	return &Room{
		Code:    code,
		clients: make(map[*Client]bool),
	}
}

// Join aggiunge un client alla stanza
func (r *Room) Join(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[c] = true
	log.Printf("👤 User %d (%s) entrato nella stanza %s — clienti: %d",
		c.UserID, c.Role, r.Code, len(r.clients))
}

// Leave rimuove un client dalla stanza
func (r *Room) Leave(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, c)
	log.Printf("👋 User %d (%s) uscito dalla stanza %s — clienti rimasti: %d",
		c.UserID, c.Role, r.Code, len(r.clients))
}

// Broadcast invia il messaggio a tutti i client della stanza tranne il mittente
func (r *Room) Broadcast(sender *Client, msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for client := range r.clients {
		if client != sender {
			select {
			case client.send <- msg:
			default:
				// Il canale è pieno → client probabilmente disconnesso
				log.Printf("⚠️  Buffer pieno per user %d, chiudo", client.UserID)
				close(client.send)
				delete(r.clients, client)
			}
		}
	}
}

// Size restituisce il numero di client connessi
func (r *Room) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}
