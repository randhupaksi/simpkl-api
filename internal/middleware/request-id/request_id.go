package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"simpkl-api/internal/shared/constants"
)

func New() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(constants.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(constants.RequestIDContextKey, requestID)
		c.Header(constants.RequestIDHeader, requestID)
		c.Next()
	}
}
