package controllers

import (
	"RunPulse/middleware"
	"RunPulse/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ResultController struct {
	DB *gorm.DB
}

func NewResultController(db *gorm.DB) *ResultController {
	return &ResultController{DB: db}
}

// POST /results
// Salva il risultato di una run
func (rc *ResultController) Save(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		SessionID   uint    `json:"session_id"   binding:"required"`
		Role        string  `json:"role"         binding:"required,oneof=start end"`
		TimeSeconds float64 `json:"time_seconds" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verifica che l'utente faccia parte della sessione
	var session models.Session
	if err := rc.DB.First(&session, req.SessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sessione non trovata"})
		return
	}

	isStart := session.StartUserID != nil && *session.StartUserID == userID
	isEnd := session.EndUserID != nil && *session.EndUserID == userID

	if !isStart && !isEnd {
		c.JSON(http.StatusForbidden, gin.H{"error": "non sei parte di questa sessione"})
		return
	}

	result := models.Result{
		SessionID:   req.SessionID,
		UserID:      userID,
		Role:        req.Role,
		TimeSeconds: req.TimeSeconds,
	}

	if err := rc.DB.Create(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "errore salvataggio risultato"})
		return
	}

	// Marca la sessione come finished
	rc.DB.Model(&session).Update("status", models.StatusFinished)

	c.JSON(http.StatusCreated, result)
}

// GET /results
// Storico risultati dell'utente autenticato
func (rc *ResultController) MyResults(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var results []models.Result
	rc.DB.Where("user_id = ?", userID).
		Preload("Session").
		Preload("Session.StartUser").
		Preload("Session.EndUser").
		Order("created_at DESC").
		Find(&results)

	c.JSON(http.StatusOK, results)
}

// GET /results/session/:session_id
// Risultati di una specifica sessione
func (rc *ResultController) BySession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID := c.Param("session_id")

	// Verifica che l'utente sia nella sessione
	var session models.Session
	if err := rc.DB.First(&session, sessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sessione non trovata"})
		return
	}

	isStart := session.StartUserID != nil && *session.StartUserID == userID
	isEnd := session.EndUserID != nil && *session.EndUserID == userID
	if !isStart && !isEnd {
		c.JSON(http.StatusForbidden, gin.H{"error": "accesso negato"})
		return
	}

	var results []models.Result
	rc.DB.Where("session_id = ?", sessionID).
		Preload("User").
		Order("created_at DESC").
		Find(&results)

	c.JSON(http.StatusOK, results)
}
