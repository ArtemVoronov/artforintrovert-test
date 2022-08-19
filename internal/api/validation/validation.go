package validation

import (
	"errors"
	"net/http"

	"github.com/ArtemVoronov/artforintrovert-test/internal/api"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ApiError struct {
	Field string
	Msg   string
}

func SendError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]ApiError, len(ve))
		for i, fe := range ve {
			out[i] = ApiError{fe.Field(), message(fe.Tag())}
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": out})
		return
	}

	c.JSON(http.StatusBadRequest, api.ERROR_MESSAGE_PARSING_BODY_JSON)
}

func message(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	}
	return ""
}
