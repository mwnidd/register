package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger logs HTTP method, path, and execution time.
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		log.Printf("%s %s %s", c.Method(), c.OriginalURL(), duration)
		return err
	}
}
