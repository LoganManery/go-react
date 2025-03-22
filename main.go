package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/loganmanery/go-react-app/db"
	"github.com/loganmanery/go-react-app/models"
	"github.com/loganmanery/go-react-app/services"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Configure database
	dbConfig := db.DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "web_application_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to database
	database, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize repositories
	userRepo := models.NewUserRepository(database.Pool)
	sessionRepo := models.NewSessionRepository(database.Pool)
	auditRepo := models.NewAuditLogRepository(database.Pool)

	// Initialize services
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	tokenExpiryMin := getEnvAsInt("TOKEN_EXPIRY_MINUTES", 60)
	authService := services.NewAuthService(database.Pool, jwtSecret, tokenExpiryMin)

	// Create admin user if not exists
	ctx := context.Background()
	createAdminUser(ctx, userRepo)

	// Start session cleanup in background
	go scheduleSessionCleanup(ctx, sessionRepo)

	// Set up HTTP server with Gin
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Create a default gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.Default())

	// Setup the React app serving
	setupViteReactApp(router)

	// Define API Routes
	setupAPIRoutes(router, authService, userRepo, sessionRepo, auditRepo)

	// Get port from environment or use default
	port := getEnv("PORT", "8080")

	// Create a server with a shutdown timeout
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine so it doesn't block the graceful shutdown handling
	go func() {
		log.Printf("Server starting on port %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupViteReactApp(router *gin.Engine) {
	router.Static("/assets", "./client/dist/assets")
	router.StaticFile("/favicon.ico", "./client/dist/favicon.ico")

	// Serve other static files that might be in the root directory
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

func setupAPIRoutes(router *gin.Engine, authService *services.AuthService, userRepo *models.UserRepository, sessionRepo *models.SessionRepository, auditRepo *models.AuditLogRepository) {
	// Group API routes
	api := router.Group("/api")
	{
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello from the go gin server",
			})
		})

		// Auth routes
		auth := api.Group("/auth")
		{
			// TODO: Add auth endpoints
			// Example:
			// auth.POST("/login", handlers.Login(authService))
			// auth.POST("/register", handlers.Register(authService))
			// auth.POST("/logout", middleware.Authenticated(), handlers.Logout(authService))
		}

		// User routes
		users := api.Group("/users")
		{
			// TODO: Add user endpoints
		}

		// TODO: Add more API endpoints as needed
	}
}

// Create admin user if it doesn't exist
func createAdminUser(ctx context.Context, userRepo *models.UserRepository) {
	adminEmail := getEnv("ADMIN_EMAIL", "admin@example.com")
	adminUsername := getEnv("ADMIN_USERNAME", "admin")
	adminPassword := getEnv("ADMIN_PASSWORD", "admin_password")

	// Check if admin exists
	admin, err := userRepo.GetByEmail(ctx, adminEmail)
	if err != nil {
		log.Printf("Error checking admin user: %v", err)
		return
	}

	// If admin doesn't exist, create one
	if admin == nil {
		admin = &models.User{
			Username:        adminUsername,
			Email:           adminEmail,
			FirstName:       "Admin",
			LastName:        "User",
			IsEmailVerified: true,
			IsActive:        true,
		}

		if err := userRepo.Create(ctx, admin, adminPassword); err != nil {
			log.Printf("Error creating admin user: %v", err)
			return
		}

		log.Printf("Admin user created with email: %s", adminEmail)
	}
}

// Schedule regular cleanup of expired sessions
func scheduleSessionCleanup(ctx context.Context, sessionRepo *models.SessionRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count, err := sessionRepo.DeleteExpiredSessions(ctx)
			if err != nil {
				log.Printf("Error cleaning up expired sessions: %v", err)
			} else if count > 0 {
				log.Printf("Cleaned up %d expired sessions", count)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Helper function to read environment variables with default values
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to read environment variables as integers with default values
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		log.Printf("Warning: Invalid value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}
