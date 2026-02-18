package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	Port       string
}

var Cfg Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env non trovato, uso variabili d'ambiente di sistema")
	}

	Cfg = Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "runpulse-user"),
		DBPassword: getEnv("DB_PASSWORD", "ChangeMe123"),
		DBName:     getEnv("DB_NAME", "runpulsedb"),
		JWTSecret:  getEnv("JWT_SECRET", "supersecret"),
		Port:       getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ConnectDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Rome",
		Cfg.DBHost, Cfg.DBPort, Cfg.DBUser, Cfg.DBPassword, Cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Connessione DB fallita: %v", err)
	}

	log.Println("✅ Connesso a PostgreSQL")
	return db
}
