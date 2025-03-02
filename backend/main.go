package main

import (
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

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
	r.GET("/session/:sessionID/answers", getAnswers)
    r.Run(":8080")
}