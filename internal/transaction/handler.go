package transaction

import (
	"bytes"
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/moninte/backend/internal/core"
)

type Handler struct {
	core.Handler
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) IngestSMS(c *fiber.Ctx) error {
	var req IngestSMSRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	raw, err := h.service.IngestSMS(h.UserID(c), req.SMSText)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(IngestSMSResponse{
		Message: "SMS received and queued for processing",
		ID:      raw.ID,
	})
}

func (h *Handler) IngestManual(c *fiber.Ctx) error {
	var req IngestManualRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	tx, err := h.service.IngestManual(h.UserID(c), req.Amount, req.Type, req.Merchant, req.Category, req.Description)
	if err != nil {
		return h.Fail(c, 400, err)
	}
	return c.Status(fiber.StatusCreated).JSON(TransactionResponse{Transaction: tx})
}

func (h *Handler) IngestSMSBatch(c *fiber.Ctx) error {
	var req IngestSMSBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return h.Fail(c, 400, err)
	}
	if len(req.Messages) == 0 {
		return h.Fail(c, 400, fiber.NewError(400, "messages array is required"))
	}
	if len(req.Messages) > 500 {
		return h.Fail(c, 400, fiber.NewError(400, "too many messages, max 500 per batch"))
	}
	processed, skipped := h.service.IngestSMSBatch(h.UserID(c), req.Messages)
	return c.Status(fiber.StatusAccepted).JSON(IngestSMSBatchResponse{
		Message:   "Batch queued for processing",
		Processed: processed,
		Skipped:   skipped,
	})
}

func (h *Handler) IngestUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return h.Fail(c, 400, fiber.NewError(400, "file is required"))
	}

	name := strings.ToLower(file.Filename)
	switch {
	case strings.HasSuffix(name, ".csv"), strings.HasSuffix(name, ".pdf"), strings.HasSuffix(name, ".xlsx"):
	default:
		return h.Fail(c, 400, fiber.NewError(400, "only CSV, PDF and XLSX files are supported"))
	}

	f, err := file.Open()
	if err != nil {
		return h.Fail(c, 500, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return h.Fail(c, 500, err)
	}

	// Verify actual MIME type against declared extension
	mime := http.DetectContentType(data)
	switch {
	case strings.HasSuffix(name, ".pdf") && !strings.Contains(mime, "pdf"):
		return h.Fail(c, 400, fiber.NewError(400, "file content does not match PDF extension"))
	case strings.HasSuffix(name, ".xlsx") &&
		!strings.Contains(mime, "zip") && // xlsx is a zip
		!strings.Contains(mime, "officedocument"):
		return h.Fail(c, 400, fiber.NewError(400, "file content does not match XLSX extension"))
	case strings.HasSuffix(name, ".csv") &&
		!strings.Contains(mime, "text") &&
		!strings.Contains(mime, "octet-stream"): // some CSVs detected as binary
		return h.Fail(c, 400, fiber.NewError(400, "file content does not match CSV extension"))
	}

	var rows [][]string

	switch {
	case strings.HasSuffix(name, ".pdf"):
		rows, err = parsePDFBytes(data)
		if err != nil {
			return h.Fail(c, 422, fiber.NewError(422, "could not extract transactions from PDF: "+err.Error()))
		}
	case strings.HasSuffix(name, ".xlsx"):
		rows, err = parseXLSXBytes(data)
		if err != nil {
			return h.Fail(c, 422, fiber.NewError(422, "could not extract transactions from XLSX: "+err.Error()))
		}
	default:
		reader := csv.NewReader(bytes.NewReader(data))
		reader.TrimLeadingSpace = true
		reader.LazyQuotes = true
		rows, err = reader.ReadAll()
		if err != nil {
			return h.Fail(c, 400, fiber.NewError(400, "could not parse CSV file"))
		}
		if len(rows) > 0 {
			rows = rows[1:]
		}
	}

	if err := validateUploadRows(rows); err != nil {
		return h.Fail(c, 400, fiber.NewError(400, err.Error()))
	}

	userID := h.UserID(c)
	go func() {
		h.service.IngestCSV(userID, rows)
	}()

	return c.Status(fiber.StatusAccepted).JSON(IngestUploadResponse{
		Message:   "File accepted and queued for processing",
		Processed: len(rows),
		Skipped:   0,
	})
}

func (h *Handler) GetTransactions(c *fiber.Ctx) error {
	page, limit := h.ParsePage(c)
	txs, err := h.service.GetTransactions(h.UserID(c), page, limit)
	if err != nil {
		return h.Fail(c, 500, err)
	}
	return c.JSON(TransactionListResponse{
		Transactions: txs,
		PageMeta:     core.PageMeta{Page: page, Limit: limit},
	})
}
