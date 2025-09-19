package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// HandleBindingError serialises binding/validation issues into a user-friendly HTTP 400 response.
func HandleBindingError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	message := bindingErrorMessage(err)
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func bindingErrorMessage(err error) string {
	switch {
	case errors.Is(err, io.EOF):
		return "request body is required"
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return fmt.Sprintf("invalid JSON at position %d", syntaxErr.Offset)
	}

	var unmarshErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshErr) {
		return fmt.Sprintf("invalid value for %s", toLowerSnake(unmarshErr.Field))
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		messages := make([]string, 0, len(validationErrs))
		for _, fe := range validationErrs {
			field := toLowerSnake(fe.Field())
			switch fe.Tag() {
			case "required":
				messages = append(messages, fmt.Sprintf("%s is required", field))
			case "email":
				messages = append(messages, fmt.Sprintf("%s must be a valid email", field))
			case "min":
				messages = append(messages, fmt.Sprintf("%s must be at least %s characters", field, fe.Param()))
			case "oneof":
				messages = append(messages, fmt.Sprintf("%s must be one of [%s]", field, strings.ReplaceAll(fe.Param(), " ", ", ")))
			default:
				messages = append(messages, fmt.Sprintf("%s is invalid", field))
			}
		}
		return strings.Join(messages, ", ")
	}

	return "invalid request body"
}

func toLowerSnake(field string) string {
	if field == "" {
		return "field"
	}

	var builder strings.Builder
	for i, r := range field {
		if unicode.IsUpper(r) && i > 0 {
			builder.WriteRune('_')
		}
		builder.WriteRune(unicode.ToLower(r))
	}
	return builder.String()
}
