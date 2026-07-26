package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubPinger struct {
	err error
}

func (p stubPinger) PingContext(context.Context) error {
	return p.err
}

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		pingError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "healthy",
			wantStatus: http.StatusOK,
			wantBody:   `"status":"ok"`,
		},
		{
			name:       "database unavailable",
			pingError:  errors.New("database unavailable"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `"code":"DATABASE_UNAVAILABLE"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

			newHealthHandler("SIMPKL API", stubPinger{err: test.pingError})(context)

			assert.Equal(t, test.wantStatus, recorder.Code)
			assert.Contains(t, recorder.Body.String(), test.wantBody)
		})
	}
}
