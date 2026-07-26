package recovery

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"simpkl-api/internal/shared/response"
)

func New(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error(
					"panic recovered",
					zap.String("panic", fmt.Sprint(recovered)),
					zap.Stack("stack"),
				)
				response.InternalError(c)
			}
		}()

		c.Next()
	}
}
