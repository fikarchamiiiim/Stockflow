package httpx

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func BindJSON(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return &APIError{
				Status:  400,
				Code:    "VALIDATION_ERROR",
				Message: strings.Join(fieldMessages(ve), ", "),
			}
		}
		return &APIError{Status: 400, Code: "INVALID_BODY", Message: "invalid JSON body"}
	}
	return nil
}

func BindQuery(c *gin.Context, dst any) error {
	if err := c.ShouldBindQuery(dst); err != nil {
		return &APIError{Status: 400, Code: "INVALID_QUERY", Message: "invalid query parameters"}
	}
	return nil
}

func fieldMessages(ve validator.ValidationErrors) []string {
	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		switch fe.Tag() {
		case "required":
			msgs = append(msgs, fe.Field()+" is required")
		case "gte":
			msgs = append(msgs, fe.Field()+" must be >= "+fe.Param())
		case "gt":
			msgs = append(msgs, fe.Field()+" must be > "+fe.Param())
		case "min":
			msgs = append(msgs, fe.Field()+" must be at least "+fe.Param())
		case "oneof":
			msgs = append(msgs, fe.Field()+" must be one of: "+fe.Param())
		default:
			msgs = append(msgs, fe.Field()+" is invalid")
		}
	}
	return msgs
}
