package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johansgiraldo/noteops/backend/internal/config"
	"github.com/johansgiraldo/noteops/backend/internal/handlers"
	"github.com/johansgiraldo/noteops/backend/internal/middleware"
	"github.com/johansgiraldo/noteops/backend/internal/repository"
	"github.com/johansgiraldo/noteops/backend/internal/service"

	_ "github.com/johansgiraldo/noteops/backend/docs" // docs generados por swag init
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           NoteOPs API
// @version         1.0
// @description     API REST de gestión de notas académicas con cálculo automático de nota definitiva, sesiones de clase en tiempo real y reserva de turnos.
// @contact.name    Johan Sebastian Giraldo Hurtado
// @license.name    Apache 2.0
// @license.url     https://www.apache.org/licenses/LICENSE-2.0.html
// @host            localhost:8080
// @BasePath        /api
// @schemes         http https
//
// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
// @description Escribe **Bearer &lt;token&gt;** usando el JWT obtenido en /api/auth/login.

func main() {
	cfg := config.Load()

	// ── Base de datos ──────────────────────────────────────────────────────────
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("✓ connected to PostgreSQL")

	// ── Capas ─────────────────────────────────────────────────────────────────
	repo := repository.New(db)
	svc := service.New(repo, db)
	hub := handlers.NewHub(repo, svc)
	h := handlers.New(repo, svc, hub, cfg.JWTSecret)

	// ── Router ────────────────────────────────────────────────────────────────
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	// ── Documentación Swagger ─────────────────────────────────────────────────
	// Solo se expone fuera de producción para no publicar el mapa de la API.
	if cfg.AppEnv != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Println("✓ Swagger UI disponible en /swagger/index.html")
	}

	// ── Rutas públicas ────────────────────────────────────────────────────────
	r.GET("/api/health", h.Health)
	r.POST("/api/auth/login", h.Login)

	// ── WebSocket ─────────────────────────────────────────────────────────────
	r.GET("/ws/session/:id", hub.ServeWS)

	// ── Rutas públicas adicionales ────────────────────────────────────────────
	r.GET("/api/sessions/active", h.GetActiveSession)
	r.GET("/api/sessions/:id/slots", h.GetSlots)
	r.POST("/api/sessions/:id/slots/:slotID/reserve", h.ReserveSlot)

	// ── Rutas protegidas ──────────────────────────────────────────────────────
	api := r.Group("/api", middleware.Auth(cfg.JWTSecret))
	{
		api.GET("/subjects", h.GetSubjects)
		api.POST("/subjects", h.CreateSubject)
		api.POST("/subjects/:id/import", h.ImportSubjectData)
		api.PATCH("/subjects/:id", h.UpdateSubject)
		api.DELETE("/subjects/:id", h.DeleteSubject)
		api.GET("/subjects/:id/students", h.GetStudentsBySubject)
		api.GET("/subjects/:id/grades", h.GetSubjectGrades)
		api.GET("/subjects/:id/final-grades", h.GetFinalGrades)
		api.POST("/subjects/:id/enroll", h.EnrollStudent)

		api.POST("/students", h.CreateStudent)
		api.PATCH("/students/:id", h.UpdateStudent)

		api.POST("/grades", h.RecordGrade)
		api.PATCH("/grades/:id/comment", h.UpdateComment)

		api.POST("/sessions", h.CreateSession)
		api.POST("/sessions/:id/activate", h.ActivateSession)
		api.POST("/sessions/:id/deactivate", h.DeactivateSession)
	}

	// ── Servidor con graceful shutdown ────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("✓ NoteOPs backend listening on :%s [%s]", cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
