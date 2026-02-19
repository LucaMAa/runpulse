package controllers

import (
	"Chrono/middleware"
	"Chrono/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RecordController struct {
	DB *gorm.DB
}

func NewRecordController(db *gorm.DB) *RecordController {
	return &RecordController{DB: db}
}

// POST /records
func (rc *RecordController) Save(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		DistanceM   int     `json:"distance_m"   binding:"required,min=1"`
		TimeSeconds float64 `json:"time_seconds" binding:"required,gt=0"`
		Notes       string  `json:"notes"`
		RecordedAt  string  `json:"recorded_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recordedAt := time.Now()
	if req.RecordedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.RecordedAt); err == nil {
			recordedAt = t
		}
	}

	record := models.Record{
		UserID:      userID,
		DistanceM:   req.DistanceM,
		TimeSeconds: req.TimeSeconds,
		Notes:       req.Notes,
		RecordedAt:  recordedAt,
	}

	if err := rc.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "errore salvataggio record"})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// GET /records
// Tutti i record dell'utente, ordinati per distanza e tempo
func (rc *RecordController) MyRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var records []models.Record
	rc.DB.Where("user_id = ?", userID).
		Order("distance_m ASC, time_seconds ASC").
		Find(&records)

	c.JSON(http.StatusOK, records)
}

// GET /records/best
// Personal best per ogni distanza
func (rc *RecordController) PersonalBests(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Prende il record migliore (tempo minore) per ogni distanza
	type BestRow struct {
		DistanceM   int     `json:"distance_m"`
		TimeSeconds float64 `json:"time_seconds"`
		RecordedAt  string  `json:"recorded_at"`
	}

	var bests []BestRow
	rc.DB.Raw(`
		SELECT DISTINCT ON (distance_m)
			distance_m, time_seconds, recorded_at
		FROM records
		WHERE user_id = ?
		ORDER BY distance_m ASC, time_seconds ASC
	`, userID).Scan(&bests)

	c.JSON(http.StatusOK, bests)
}

// DELETE /records/:id
func (rc *RecordController) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	result := rc.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Record{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "record non trovato"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "record eliminato"})
}
