package response

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Meta struct {
	StatusCode int                    `json:"statusCode"`
	Message    string                 `json:"message"`
	Status     string                 `json:"status"`
	Timestamp  string                 `json:"timestamp"`
	RequestID  string                 `json:"requestId,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func Success(c *fiber.Ctx, data interface{}, message string, code int) error {
	if message == "" {
		message = "OK"
	}
	if code == 0 {
		code = http.StatusOK
	}
	return c.Status(code).JSON(fiber.Map{
		"meta": Meta{
			StatusCode: code,
			Message:    message,
			Status:     "success",
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		},
		"data": data,
	})
}

type ErrorBody struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Suggestion string      `json:"suggestion,omitempty"`
}

func Error(c *fiber.Ctx, code int, apiCode string, message string) error {
	if apiCode == "" {
		apiCode = "ERROR"
	}
	return c.Status(code).JSON(fiber.Map{
		"meta": Meta{
			StatusCode: code,
			Message:    message,
			Status:     "error",
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		},
		"error": ErrorBody{Code: apiCode, Message: message},
	})
}

func List(c *fiber.Ctx, data interface{}, message string, pagination map[string]interface{}) error {
	meta := Meta{
		StatusCode: http.StatusOK,
		Message:    message,
		Status:     "success",
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}
	body := fiber.Map{"meta": meta, "data": data}
	if pagination != nil {
		body["pagination"] = pagination
	}
	return c.Status(http.StatusOK).JSON(body)
}
