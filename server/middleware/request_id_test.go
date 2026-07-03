package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID_SetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("requestID")
		c.String(http.StatusOK, id.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Error("requestID should not be empty")
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestRequestID_ReusesExistingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("requestID")
		c.String(http.StatusOK, id.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "custom-id-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Body.String() != "custom-id-123" {
		t.Errorf("expected 'custom-id-123', got '%s'", w.Body.String())
	}
}

func TestRequestID_GeneratesUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("requestID")
		c.String(http.StatusOK, id.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	id := w.Body.String()
	// UUID v4 is 36 chars with 4 hyphens
	if len(id) != 36 {
		t.Errorf("expected UUID length 36, got %d: '%s'", len(id), id)
	}
}
