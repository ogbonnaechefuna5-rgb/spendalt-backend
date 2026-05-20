package core

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/lang"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// AppError is a user-facing error with an associated HTTP status code.
type AppError struct {
	Status  int
	Message string
}

func (e *AppError) Error() string { return e.Message }

// NewError creates an AppError for client-facing messages.
func NewError(status int, msg string) error { return &AppError{Status: status, Message: msg} }

// isInternalError returns true for errors that should never be shown to users.
func isInternalError(err error) bool {
	msg := err.Error()
	return strings.HasPrefix(msg, "pq:") ||
		strings.HasPrefix(msg, "sql:") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host")
}

// Handler provides shared helpers all domain handlers embed.
type Handler struct{}

func (h *Handler) UserID(c *fiber.Ctx) string {
	id, _ := c.Locals("user_id").(string)
	return id
}

// ParsePage parses and clamps page/limit query params.
// Defaults: page=1, limit=20. Max limit=100.
func (h *Handler) ParsePage(c *fiber.Ctx) (page, limit int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	limit, _ = strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return
}

func (h *Handler) OK(c *fiber.Ctx, key string, data interface{}) error {
	return c.JSON(fiber.Map{key: data})
}

func (h *Handler) Created(c *fiber.Ctx, key string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{key: data})
}

// Fail writes an error response. It respects AppError for status+message,
// maps ErrNotFound to 404, masks any internal/DB error as 500,
// and passes through all other errors at the given status.
func (h *Handler) Fail(c *fiber.Ctx, status int, err error) error {
	rid := h.RequestID(c)
	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Status).JSON(ErrorResponse{Error: appErr.Message, RequestID: rid})
	}
	if errors.Is(err, ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: lang.ErrNotFound, RequestID: rid})
	}
	if status >= 500 || isInternalError(err) {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: lang.ErrInternal, RequestID: rid})
	}
	return c.Status(status).JSON(ErrorResponse{Error: err.Error(), RequestID: rid})
}

func (h *Handler) RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("request_id").(string)
	return id
}

func (h *Handler) Message(c *fiber.Ctx, msg string) error {
	return c.JSON(MessageResponse{Message: msg})
}
