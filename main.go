package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"register/adapter/api/handler"
	"register/adapter/repository"
	"register/config"
	"register/core/services"
	"register/pkg/middleware"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var cfg config.Config
	if err := config.ReadConfig("config/config.yml", &cfg); err != nil {
		log.Fatal("Cannot load config:", err)
	}

	log.SetFlags(0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		log.Fatal("Cannot connect to MongoDB:", err)
	}
	db := client.Database(cfg.Mongo.DBName)

	userRepo := repository.NewMongoRepository(db)
	userService := services.NewUserService(userRepo, cfg.App.JWTSecret)
	userHandler := handler.NewUserHandler(userService)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(middleware.Logger())

	// Public Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Post("/register", userHandler.Register)
	app.Post("/login", userHandler.Login)

	// Private Routes (Group & Middleware)
	api := app.Group("/api", middleware.Auth(cfg.App.JWTSecret))
	api.Get("/users", userHandler.List)
	api.Get("/users/:id", userHandler.Get)
	api.Put("/users/:id", userHandler.Update)
	api.Delete("/users/:id", userHandler.Delete)

	for _, routes := range app.Stack() {
		for _, r := range routes {
			logJSON("INFO", fmt.Sprintf("[Server] %s %s", r.Method, r.Path))
		}
	}

	serverErr := make(chan error, 1)
	go func() {
		logJSON("INFO", fmt.Sprintf("[Server] Start on port: %s", cfg.Server.Port))
		serverErr <- app.Listen(cfg.Server.Port)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		if err != nil {
			logJSON("ERROR", fmt.Sprintf("[Server] Listen error: %v", err))
			return
		}
	case <-quit:
	}

	log.Println("Shutting down server...")

	// Fiber Shutdown
	if err := app.Shutdown(); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	// Drain listen result if we triggered shutdown first.
	select {
	case <-serverErr:
	default:
	}

	if err := client.Disconnect(context.Background()); err != nil {
		log.Fatal("Error disconnecting from MongoDB:", err)
	}

	log.Println("Server exited properly")
}

func logJSON(severity, message string) {
	entry := map[string]string{
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"severity":  severity,
		"message":   message,
	}
	if data, err := json.Marshal(entry); err == nil {
		log.Println(string(data))
		return
	}
	log.Printf("%s %s", severity, message)
}
