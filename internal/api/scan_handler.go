package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sekai/internal/api/apierrs"
	"sekai/internal/entities/domain"
	"sekai/internal/entities/dto"
	"sekai/internal/usecase/scan"
	"sekai/pkg/ctxutils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ScanService interface {
	Create(ctx context.Context, cmd scan.StartCommand) (domain.Scan, error)
	ListByUser(ctx context.Context, userID int64) ([]domain.Scan, error)
}
type ScanHandler struct {
	svc ScanService
}

func NewScanHandler(svc ScanService) *ScanHandler {
	return &ScanHandler{svc: svc}
}

func (h *ScanHandler) StartScan(c *gin.Context) {
	log := ctxutils.Logger(c.Request.Context())
	user, ok := ctxutils.User(c.Request.Context())
	if !ok {
		log.Error("user not found in context", "method", "StartScan")
		c.Error(apierrs.ErrUnauthorized("no user provided"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Error("invalid file", "error", err)
		c.Error(apierrs.ErrBadRequest("invalid file"))
		return
	}

	filename := filepath.Base(file.Filename)
	if filename == "." || filename == string(filepath.Separator) || filename == "" {
		c.Error(apierrs.ErrBadRequest("invalid file name"))
		return
	}
	if strings.ToLower(filepath.Ext(filename)) != ".zip" {
		c.Error(apierrs.ErrBadRequest("only zip archives are supported"))
		return
	}

	tmpDir, err := os.MkdirTemp("", "uploads-*")
	if err != nil {
		log.Error("failed to create temp dir", "error", err)
		c.Error(apierrs.ErrInternalServerError())
		return
	}
	defer os.RemoveAll(tmpDir)

	dst := filepath.Join(tmpDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		log.Error("failed to save uploaded file", "error", err)
		c.Error(apierrs.ErrInternalServerError())
		return
	}

	enableSAST, err := parseBoolForm(c, "enable_sast", true)
	if err != nil {
		log.Error("invalid parsing enable_sast", "error", err)
		c.Error(apierrs.ErrBadRequest("invalid enable_sast parameter"))
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.Error(apierrs.ErrBadRequest("invalid multipart form"))
		return
	}
	labels := append([]string{}, form.Value["labels"]...)
	labels = append(labels, form.Value["labels[]"]...)
	if len(labels) > 3 {
		c.Error(apierrs.ErrBadRequest("labels count must not exceed 3"))
		return
	}
	request := dto.CreateScanRequest{
		EnableSAST: enableSAST,
		Labels:     normalizeLabels(labels),
	}

	cmd := scan.StartCommand{
		EnableSAST:  request.EnableSAST,
		SizeBytes:   file.Size,
		UserID:      user.ID,
		ArchivePath: dst,
		Labels:      request.Labels,
	}

	createdScan, err := h.svc.Create(c.Request.Context(), cmd)
	if err != nil {
		log.Error("failed to create scan", "error", err)
		handleScanError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.NewScanResponse(createdScan))
}

func (h *ScanHandler) ListScans(c *gin.Context) {
	log := ctxutils.Logger(c.Request.Context())
	user, ok := ctxutils.User(c.Request.Context())
	if !ok {
		log.Error("user not found in context", "method", "ListScans")
		c.Error(apierrs.ErrUnauthorized("no user provided"))
		return
	}

	scans, err := h.svc.ListByUser(c.Request.Context(), user.ID)
	if err != nil {
		log.Error("failed to list scans", "error", err)
		handleScanError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewListScansResponse(scans))
}

func parseBoolForm(c *gin.Context, key string, defaultValue bool) (bool, error) {

	raw := c.PostForm(key)
	if raw == "" {
		return defaultValue, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("invalid bool field %q", key)
	}
	return v, nil
}

func normalizeLabels(labels []string) []string {
	normalized := make([]string, 0, len(labels))
	for _, label := range labels {
		normalized = append(normalized, strings.TrimSpace(label))
	}
	return normalized
}

func handleScanError(c *gin.Context, err error) {
	var domainErr domain.Error

	switch {
	case errors.Is(err, scan.ErrScanQuotaExceeded):
		c.Error(apierrs.ErrForbidden("scan quota exceeded"))
	case errors.Is(err, scan.ErrInvalidArtifact):
		c.Error(apierrs.ErrBadRequest("invalid artifact"))
	case errors.As(err, &domainErr):
		c.Error(apierrs.ErrBadRequest(domainErr.Message))
	default:
		c.Error(apierrs.ErrInternalServerError())
	}
}
