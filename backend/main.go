package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"sync"
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
	mu        sync.Mutex
}

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex

func createSession(c *gin.Context) {
	sessionID := uuid.New().String()
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


func main() {
	r := gin.Default()
	r.POST("/session", createSession)
	r.POST("/session/:sessionID/join", joinSession)
	r.GET("/sessions", getSessions)
	r.GET("/session/:sessionID", getSession)
	r.Run(":8080")
}
