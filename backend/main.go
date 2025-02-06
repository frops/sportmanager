package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func main() {
	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Миграция базы данных
	db.AutoMigrate(&Match{}, &Player{})

	// Инициализация Telegram бота
	bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal("Failed to initialize Telegram bot:", err)
	}

	// Настройка Gin роутера
	r := gin.Default()

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
		// Матчи
		api.GET("/matches", getMatches)
		api.POST("/matches", createMatch)
		api.POST("/matches/:id/join", joinMatch)
		api.DELETE("/matches/:id/leave", leaveMatch)
		api.DELETE("/matches/:id", deleteMatch)
		api.POST("/matches/:id/restore", restoreMatch)

		// Игроки
		api.GET("/players", getPlayers)
		api.POST("/players", createPlayer)
	}

	// Запуск Telegram бота в отдельной горутине
	go runTelegramBot()

	// Запуск сервера
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func runTelegramBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Обработка команд
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Welcome to Football Match Manager! Use /help for available commands.")
				bot.Send(msg)
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					`Available commands:
/matches - show upcoming matches
/join [ID] - join a match
/leave [ID] - leave a match`)
				bot.Send(msg)
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

	// Просто отмечаем матч как неактивный
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
