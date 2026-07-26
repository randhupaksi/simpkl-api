package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"simpkl-api/internal/shared/constants"
)

type Envelope struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Data      any                 `json:"data,omitempty"`
	Meta      any                 `json:"meta"`
	Code      string              `json:"code,omitempty"`
	Errors    map[string][]string `json:"errors,omitempty"`
	RequestID string              `json:"request_id,omitempty"`
}

func Success(
	c *gin.Context,
	status int,
	message string,
	data any,
	meta any,
) {
	c.JSON(status, Envelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Error(
	c *gin.Context,
	status int,
	message string,
	code string,
	errors map[string][]string,
) {
	requestID, _ := c.Get(constants.RequestIDContextKey)

	c.AbortWithStatusJSON(status, Envelope{
		Success:   false,
		Message:   message,
		Meta:      nil,
		Code:      code,
		Errors:    errors,
		RequestID: stringValue(requestID),
	})
}

func InternalError(c *gin.Context) {
	Error(
		c,
		http.StatusInternalServerError,
		"Terjadi kesalahan pada server",
		"INTERNAL_SERVER_ERROR",
		nil,
	)
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}
