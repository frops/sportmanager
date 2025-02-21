package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sportmanager/internal/config"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Match struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Date       time.Time `json:"date"`
	Location   string    `json:"location"`
	VenueName  string    `json:"venueName" gorm:"default:Nova Sports Soccer Field"`
	MapLink    string    `json:"mapLink"`
	MinPlayers int       `json:"minPlayers" gorm:"default:10"`
	MaxPlayers int       `json:"maxPlayers" gorm:"default:12"`
	Players    []Player  `json:"players" gorm:"many2many:match_players;"`
	Active     bool      `json:"active" gorm:"default:true"`
}

type Player struct {
	ID         uint    `json:"id" gorm:"primaryKey"`
	TelegramID int64   `json:"telegramId"`
	Name       string  `json:"name"`
	Matches    []Match `json:"matches" gorm:"many2many:match_players;"`
}

type JoinMatchRequest struct {
	Name string `json:"name" binding:"required"`
}

var db *gorm.DB
var bot *tgbotapi.BotAPI
var logger *zap.Logger

func init() {
	var err error
	logConfig := zap.NewProductionConfig()

	if os.Getenv("ENV") == config.EnvProduction {
		logConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logger, err = logConfig.Build()
	} else {
		logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		logger, err = logConfig.Build()
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

func runMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	logger.Info("Applied migrations", zap.Int("count", n))
	return nil
}

func main() {
	// Database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		os.Getenv("PGHOST"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGPORT"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize Telegram bot only if no other instance is running
	if os.Getenv("ENABLE_BOT") == "true" {
		bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
		if err != nil {
			logger.Fatal("Failed to initialize Telegram bot", zap.Error(err))
		}
		// Start Telegram bot in a separate goroutine
		go runTelegramBot()
	}

	// Setup Gin router
	if os.Getenv("ENV") == config.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Health check endpoint
	r.GET(config.HealthCheckPath, func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Failed to get database instance", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "reason": "database error"})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			logger.Error("Failed to ping database", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "reason": "database connection error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API endpoints
	api := r.Group("/api")
	{
		// Matches
		api.GET("/matches", getMatches)
		api.POST("/matches", createMatch)
		api.POST("/matches/:id/join", joinMatch)
		api.DELETE("/matches/:id/leave", leaveMatch)
		api.DELETE("/matches/:id", deleteMatch)
		api.POST("/matches/:id/restore", restoreMatch)

		// Players
		api.GET("/players", getPlayers)
		api.POST("/players", createPlayer)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    config.DefaultServerPort,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Close database connection
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Error("Error closing database connection", zap.Error(err))
		}
	}

	// Stop server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

func runTelegramBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Handle commands
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Welcome to Sport Manager! Use /help for available commands.")
				if _, err := bot.Send(msg); err != nil {
					logger.Error("Failed to send telegram message", zap.Error(err))
				}
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					`Available commands:
/matches - show upcoming matches
/join [ID] - join a match
/leave [ID] - leave a match`)
				if _, err := bot.Send(msg); err != nil {
					logger.Error("Failed to send telegram message", zap.Error(err))
				}
			}
		}
	}
}

// API handlers
func getMatches(c *gin.Context) {
	var matches []Match
	if err := db.Preload("Players").Find(&matches).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch matches"})
		return
	}
	c.JSON(200, matches)
}

func createMatch(c *gin.Context) {
	var match Match
	if err := c.BindJSON(&match); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Set default values if not provided
	if match.VenueName == "" {
		// Try to get the last match's venue
		var lastMatch Match
		if err := db.Order("created_at DESC").First(&lastMatch).Error; err == nil {
			match.VenueName = lastMatch.VenueName
		} else {
			match.VenueName = "Nova Sports Soccer Field"
		}
	}

	if match.MinPlayers == 0 {
		match.MinPlayers = 10
	}

	if match.MaxPlayers == 0 {
		match.MaxPlayers = 12
	}

	match.Active = true

	if err := db.Create(&match).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create match"})
		return
	}

	c.JSON(201, match)
}

func joinMatch(c *gin.Context) {
	var req JoinMatchRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	matchID := c.Param("id")
	var match Match
	if err := db.Preload("Players").First(&match, matchID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Match not found"})
		return
	}

	if len(match.Players) >= match.MaxPlayers {
		c.JSON(400, gin.H{"error": "Match is full"})
		return
	}

	var player Player
	result := db.Where("name = ?", req.Name).FirstOrCreate(&player, Player{Name: req.Name})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to create player"})
		return
	}

	if err := db.Model(&match).Association("Players").Append(&player); err != nil {
		c.JSON(500, gin.H{"error": "Failed to join match"})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully joined match"})
}

func leaveMatch(c *gin.Context) {
	matchID := c.Param("id")
	var match Match
	if err := db.First(&match, matchID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Match not found"})
		return
	}

	var req JoinMatchRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	var player Player
	if err := db.Where("name = ?", req.Name).First(&player).Error; err != nil {
		c.JSON(404, gin.H{"error": "Player not found"})
		return
	}

	if err := db.Model(&match).Association("Players").Delete(&player); err != nil {
		c.JSON(500, gin.H{"error": "Failed to leave match"})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully left match"})
}

func deleteMatch(c *gin.Context) {
	matchID := c.Param("id")
	var match Match
	if err := db.First(&match, matchID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Match not found"})
		return
	}

	// Just mark the match as inactive
	if err := db.Model(&match).Update("active", false).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to cancel match"})
		return
	}

	c.JSON(200, gin.H{"message": "Match successfully cancelled"})
}

func restoreMatch(c *gin.Context) {
	matchID := c.Param("id")
	var match Match
	if err := db.First(&match, matchID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Match not found"})
		return
	}

	if err := db.Model(&match).Update("active", true).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to restore match"})
		return
	}

	c.JSON(200, gin.H{"message": "Match successfully restored"})
}

func getPlayers(c *gin.Context) {
	var players []Player
	if err := db.Find(&players).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch players"})
		return
	}
	c.JSON(200, players)
}

func createPlayer(c *gin.Context) {
	var player Player
	if err := c.BindJSON(&player); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	if err := db.Create(&player).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create player"})
		return
	}

	c.JSON(201, player)
}
