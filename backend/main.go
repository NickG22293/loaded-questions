package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"math/rand"
	"time"
	"encoding/json"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GameSession struct {
	ID        string            `json:"id"`
	Asker     *Player           `json:"asker"`
	Players   map[string]Player `json:"players"`
	Question  string            `json:"question"`
	Answers   map[string]string `json:"answers"`
	Guesses   map[string]string `json:"guesses"`
	Score     map[string]int    `json:"score"`
	Started   bool              `json:"started"`
	mu        sync.Mutex
}

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex

func createSession(c *gin.Context) {
	sessionID := uuid.New().String()[:6]
	sessionsMu.Lock()
	sessions[sessionID] = &GameSession{
		ID:      sessionID,
		Players: make(map[string]Player),
		Answers: make(map[string]string),
		Guesses: make(map[string]string),
		Score:   make(map[string]int),
	}
	sessionsMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"session_id": sessionID})
}

func startGame(c *gin.Context) {
	sessionID := c.Param("sessionID")

	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	sessionsMu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()
	if len(session.Players) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least 3 players required to start"})
		return
	}

	// Randomly select one player to be the Asker
	rand.Seed(time.Now().UnixNano())
    playerIDs := make([]string, 0, len(session.Players))
    for playerID := range session.Players {
        playerIDs = append(playerIDs, playerID)
    }
    randomIndex := rand.Intn(len(session.Players))
	asker := session.Players[playerIDs[randomIndex]] 
    session.Asker = &asker

	session.Started = true

	c.JSON(http.StatusOK, gin.H{"message": "Game started!"})
}

func joinSession(c *gin.Context) {
	sessionID := c.Param("sessionID")
	playerName := c.Query("name")
	if playerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	sessionsMu.Unlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	playerID := uuid.New().String()
	player := Player{ID: playerID, Name: playerName}

	session.mu.Lock()
	session.Players[playerID] = player
	if session.Asker == nil {
		session.Asker = &player
	}
	session.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"player_id": playerID, "session_id": sessionID})
}

func getSessions(c *gin.Context) {
	jsonData, err := json.Marshal(sessions)
	if err != nil {

		c.JSON(http.StatusInternalServerError, "Problem fetching sessions")
	}
	c.JSON(http.StatusFound, jsonData)
}

func getSession(c *gin.Context) {
	sessionID := c.Param("sessionID")

	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	sessionsMu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func setQuestion(c *gin.Context) {
	sessionID := c.Param("sessionID")

	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	sessionsMu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var requestBody struct {
		Question string `json:"question"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	session.mu.Lock()
	session.Question = requestBody.Question
	session.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "Question set!"})
}

func submitAnswer(c *gin.Context) {
	sessionID := c.Param("sessionID")

	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	sessionsMu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var requestBody struct {
		PlayerID string `json:"player_id"`
		Answer   string `json:"answer"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	session.mu.Lock()
	session.Answers[requestBody.PlayerID] = requestBody.Answer
	session.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "Answer submitted!"})
}

func main() {
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5173"}, // Allow frontend origin
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/session", createSession)
	r.POST("/session/:sessionID/join", joinSession)
	r.GET("/sessions", getSessions)
	r.GET("/session/:sessionID", getSession)
	r.POST("/session/:sessionID/question", setQuestion)
	r.POST("/session/:sessionID/answer", submitAnswer)
	r.POST("/session/:sessionID/start", startGame)
	r.Run(":8080")
}
