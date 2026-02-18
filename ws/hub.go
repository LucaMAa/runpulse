package ws

import (
	"log"
	"sync"
)

// Hub gestisce tutte le stanze WebSocket attive
type Hub struct {
	rooms map[string]*Room // code → Room
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

// GetOrCreateRoom restituisce la stanza esistente o ne crea una nuova
func (h *Hub) GetOrCreateRoom(code string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[code]; ok {
		return room
	}

	room := newRoom(code)
	h.rooms[code] = room
	log.Printf("🏠 Nuova stanza creata: %s", code)
	return room
}

// RemoveRoom elimina la stanza quando è vuota
func (h *Hub) RemoveRoom(code string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, code)
	log.Printf("🗑️  Stanza rimossa: %s", code)
}

// RoomCount restituisce il numero di stanze attive (utile per debug/monitoring)
func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}
