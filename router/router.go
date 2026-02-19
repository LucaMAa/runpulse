package router

import (
	"Chrono/controllers"
	"Chrono/middleware"
	"Chrono/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, hub *ws.Hub) *gin.Engine {
	r := gin.Default()

	// Controllers
	authCtrl   := controllers.NewAuthController(db)
	sessionCtrl := controllers.NewSessionController(db)
	resultCtrl := controllers.NewResultController(db)
	recordCtrl := controllers.NewRecordController(db)

	// ─── Public routes ───────────────────────────────────────────
	auth := r.Group("/auth")
	{
		auth.POST("/register", authCtrl.Register)
		auth.POST("/login",    authCtrl.Login)
	}

	// ─── Protected routes ────────────────────────────────────────
	api := r.Group("/")
	api.Use(middleware.AuthRequired())
	{
		// Auth
		api.GET("/auth/me", authCtrl.Me)

		// Sessions
		api.POST("/sessions/create",       sessionCtrl.Create)
		api.POST("/sessions/join",         sessionCtrl.Join)
		api.GET("/sessions/my",            sessionCtrl.My)
		api.GET("/sessions/:code",         sessionCtrl.Get)

		// Results (run in sessione)
		api.POST("/results",                      resultCtrl.Save)
		api.GET("/results",                       resultCtrl.MyResults)
		api.GET("/results/session/:session_id",   resultCtrl.BySession)

		// Records (personal bests)
		api.POST("/records",        recordCtrl.Save)
		api.GET("/records",         recordCtrl.MyRecords)
		api.GET("/records/best",    recordCtrl.PersonalBests)
		api.DELETE("/records/:id",  recordCtrl.Delete)
	}

	// ─── WebSocket ───────────────────────────────────────────────
	r.GET("/ws", ws.ServeWS(hub))

	return r
}
