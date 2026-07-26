package errorhandler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"simpkl-api/internal/shared/response"
)

func New(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		lastError := c.Errors.Last()
		log.Error("unhandled request error", zap.Error(lastError.Err))
		response.InternalError(c)
	}
}
