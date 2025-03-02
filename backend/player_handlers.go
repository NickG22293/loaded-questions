package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
		PlayerID PlayerID `json:"player_id"`
		Answer   string   `json:"answer"`
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
	session.Status = Answering
	session.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "Question set!"})
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

	playerID := PlayerID(uuid.New().String())
	player := Player{ID: playerID, Name: playerName}

	session.mu.Lock()
	session.Players[playerID] = player
	if session.Asker == nil {
		session.Asker = &player
	}
	session.mu.Unlock()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}

	session.mu.Lock()
	session.playerConnections[playerID] = conn
	session.mu.Unlock()

	go handlePlayerConnection(session, playerID, conn)

	c.JSON(http.StatusOK, gin.H{"player_id": playerID, "session_id": sessionID})
}