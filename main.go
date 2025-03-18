package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Create a default gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.Default())

	// Setup the React app serving
	setupViteReactApp(router)

	// Define API Routes
	setupAPIRoutes(router)

	// Get port from enviroment or use default
	port := getEnvOrDefault("PORT", "8080")

	// Start the server
	log.Printf("Server starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func setupViteReactApp(router *gin.Engine) {
	router.Static("/assets", "./client/dist/assets")
	router.StaticFile("/favicon.ico", "./client/dist/favicon.ico")

	// Server other static files that might be in the root directory
	fs := http.Dir("./client/dist")
	fileServer := http.StripPrefix("/", http.FileServer(fs))
	router.GET("/vite.svg", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// For any other routes, serve the index.html file
	router.NoRoute(func(c *gin.Context) {
		// If the request is for an API endpoint, return 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		c.File("./client/dist/index.html")
	})
}

func setupAPIRoutes(router *gin.Engine) {
	// Group API routes
	api := router.Group("/api")
	{
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello from the go gin server",
			})
		})
		// TODO: Add more API endpoints as needed
	}
}

// getEnvOrDefault returnsd an enviroment variable or a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
