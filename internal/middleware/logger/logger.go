package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"simpkl-api/internal/shared/constants"
)

func New(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		requestID, _ := c.Get(constants.RequestIDContextKey)
		fields := []zap.Field{
			zap.String("request_id", stringValue(requestID)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int("response_size", c.Writer.Size()),
			zap.Duration("duration", time.Since(startedAt)),
			zap.String("client_ip", c.ClientIP()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		log.Info("http request completed", fields...)
	}
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}
