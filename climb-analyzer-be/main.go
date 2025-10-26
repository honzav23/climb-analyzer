package main

import (
	"climb-analyzer-be/climbAnalyzer"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func recoveryMiddleware(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error"})
		}
	}()
	c.Next()
}

func main() {
	isProduction := false
	backendUrl := "localhost:8080"

	_, defined := os.LookupEnv("PRODUCTION")
	if defined {
		isProduction = true
	}

	if isProduction {
		gin.SetMode(gin.ReleaseMode)
		backendUrl = "0.0.0.0:8080"
	}

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:4200", "http://localhost"}
	config.AllowMethods = []string{"POST", "OPTIONS"}
	router.Use(cors.New(config))

	router.Use(recoveryMiddleware)
	router.POST("/api/analyze", climbAnalyzer.AnalyzeClimbs)
	router.Run(backendUrl)
}
