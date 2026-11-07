package httpx

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

func WriteError(c *gin.Context, status int, code, message string, extra map[string]any) {
	body := gin.H{"code": code}
	if message != "" {
		body["message"] = message
	}
	for k, v := range extra {
		body[k] = v
	}

	c.AbortWithStatusJSON(status, gin.H{"error": body})
}

type APIError struct {
	Status  int
	Code    string
	Message string
	Extra   map[string]any
}

func (e *APIError) Error() string {
	return e.Code
}

func Fail(c *gin.Context, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		log.Printf("internal error: %v", err)
		WriteError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
		return
	}

	WriteError(c, apiErr.Status, apiErr.Code, apiErr.Message, apiErr.Extra)
}
