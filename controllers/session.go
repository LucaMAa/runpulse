package controllers

import (
	"Chrono/middleware"
	"Chrono/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SessionController struct {
	DB *gorm.DB
}

func NewSessionController(db *gorm.DB) *SessionController {
	return &SessionController{DB: db}
}

// POST /sessions/create
// Lo START crea la sessione e ottiene il codice da condividere
func (sc *SessionController) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Genera codice univoco
	var code string
	for {
		code = models.GenerateCode()
		var existing models.Session
		if err := sc.DB.Where("code = ?", code).First(&existing).Error; err != nil {
			break // codice non esiste ancora
		}
	}

	session := models.Session{
		Code:        code,
		Status:      models.StatusWaiting,
		StartUserID: &userID,
	}

	if err := sc.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "errore creazione sessione"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.ID,
		"code":       session.Code,
		"status":     session.Status,
		"message":    "Condividi il codice con l'utente END",
	})
}

// POST /sessions/join
// L'END entra nella sessione tramite codice
func (sc *SessionController) Join(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.Session
	if err := sc.DB.Where("code = ?", req.Code).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sessione non trovata"})
		return
	}

	if session.Status != models.StatusWaiting {
		c.JSON(http.StatusConflict, gin.H{"error": "sessione già attiva o terminata"})
		return
	}

	// Evita che lo stesso utente faccia sia START che END
	if session.StartUserID != nil && *session.StartUserID == userID {
		c.JSON(http.StatusConflict, gin.H{"error": "non puoi unirti alla tua stessa sessione"})
		return
	}

	session.EndUserID = &userID
	session.Status = models.StatusActive

	if err := sc.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "errore aggiornamento sessione"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": session.ID,
		"code":       session.Code,
		"status":     session.Status,
		"message":    "Connesso alla sessione come END",
	})
}

// GET /sessions/:code
// Info su una sessione (utile per polling lato client)
func (sc *SessionController) Get(c *gin.Context) {
	code := c.Param("code")

	var session models.Session
	if err := sc.DB.Preload("StartUser").Preload("EndUser").
		Where("code = ?", code).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sessione non trovata"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GET /sessions/my
// Tutte le sessioni dell'utente autenticato
func (sc *SessionController) My(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var sessions []models.Session
	sc.DB.Where("start_user_id = ? OR end_user_id = ?", userID, userID).
		Preload("StartUser").
		Preload("EndUser").
		Preload("Results").
		Order("created_at DESC").
		Find(&sessions)

	c.JSON(http.StatusOK, sessions)
}
