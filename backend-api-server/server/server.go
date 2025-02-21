package server

import (
	"context"
	"fmt"
	loggo "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Server struct {
	cfg    *Config
	router *mux.Router
	db     *gorm.DB
}

func New(cfg *Config) *Server {
	setLogConfigFromEnv()
	s := Server{cfg: cfg}
	s.router = mux.NewRouter()
	s.setRoutes()
	s.initDB()

	return &s
}

func (s *Server) setRoutes() {
	s.router.HandleFunc("/tasks", s.handleCreateTask).Methods(http.MethodPost)
	s.router.HandleFunc("/tasks", s.handleListTasks).Methods(http.MethodGet)
	s.router.HandleFunc("/tasks/pick", s.handlePickTask).Methods(http.MethodGet)
	s.router.HandleFunc("/tasks/{id}", s.handleGetTask).Methods(http.MethodGet)
	s.router.HandleFunc("/tasks/{id}/finish", s.handleFinishTask).Methods(http.MethodPost)
}

func (s *Server) initDB() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.cfg.DBHost, s.cfg.DBPort, s.cfg.DBUser, s.cfg.DBPassword, s.cfg.DBName)
	newLogger := logger.New(
		loggo.New(os.Stdout, "\r\n", loggo.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      true,
		},
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&TaskData{})
	log.Info("Postgres connection successful")
	s.db = db
}

func setLogConfigFromEnv() {
	level, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = log.DebugLevel
	}
	log.SetLevel(level)
}

func (s *Server) Run() {
	srv := &http.Server{
		Addr:         ":" + s.cfg.ServerPort,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("Starting the server on :" + s.cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("ListenAndServe error: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Info("Server gracefully stopped")
}
