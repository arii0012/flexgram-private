
package handlers

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()

		c.Next()

		log.Printf(
			"method=%s path=%s status=%d duration=%s client=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(started),
			c.ClientIP(),
		)
	}
}
