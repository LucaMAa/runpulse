package main

import (
	"RunPulse/config"
	"RunPulse/models"
	"RunPulse/router"
	"RunPulse/ws"
	"log"
)

func main() {
	// Carica configurazione da .env
	config.Load()

	// Connetti a PostgreSQL
	db := config.ConnectDB()

	// Auto-migra i modelli (crea/aggiorna le tabelle)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Result{},
		&models.Record{},
	); err != nil {
		log.Fatalf("❌ AutoMigrate fallito: %v", err)
	}
	log.Println("✅ Schema DB aggiornato")

	// Crea il WebSocket Hub
	hub := ws.NewHub()

	// Setup router Gin
	r := router.Setup(db, hub)

	// Avvia server
	addr := ":" + config.Cfg.Port
	log.Printf("🚀 Server in ascolto su %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Server errore: %v", err)
	}
}
