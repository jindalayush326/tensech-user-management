package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxRequestID = "request_id"

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := uuid.NewString()

		c.Set(CtxRequestID, requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		duration := time.Since(start)

		log.Printf(
			"request_id=%s method=%s path=%s status=%d duration=%s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}