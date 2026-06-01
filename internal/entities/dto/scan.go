package dto

import (
	"fmt"
	"sekai/internal/entities/domain"
	"time"
)

type CreateScanRequest struct {
	Labels     []string `form:"labels"`
	EnableSAST bool     `form:"enable_sast"`
}

type ListScansResponse struct {
	Scans []ScanResponse `json:"scans"`
}

type ScanResponse struct {
	ID               int64                    `json:"id"`
	OwnerID          int64                    `json:"owner_id"`
	Status           string                   `json:"status"`
	APIType          string                   `json:"api_type"`
	ArtifactMetadata ArtifactMetadataResponse `json:"artifact_metadata"`
	Labels           []string                 `json:"labels"`
	ScanChecks       []ScanCheckResponse      `json:"scan_checks"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

type ArtifactMetadataResponse struct {
	FileName         string `json:"file_name"`
	TargetLanguage   string `json:"target_language"`
	SchemaVersion    int    `json:"schema_version"`
	SLOC             int    `json:"sloc"`
	UncompressedSize int64  `json:"uncompressed_size"`
	CompressedSize   int64  `json:"compressed_size"`
}

type ScanCheckResponse struct {
	ID         int64      `json:"id"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func NewListScansResponse(scans []domain.Scan) ListScansResponse {
	response := ListScansResponse{Scans: make([]ScanResponse, 0, len(scans))}
	for _, scan := range scans {
		response.Scans = append(response.Scans, NewScanResponse(scan))
	}
	return response
}

func NewScanResponse(scan domain.Scan) ScanResponse {
	metadata := scan.Metadata()

	labels := make([]string, 0, len(scan.Labels()))
	for _, label := range scan.Labels() {
		labels = append(labels, label.Name())
	}

	checks := make([]ScanCheckResponse, 0, len(scan.ScanChecks()))
	for _, check := range scan.ScanChecks() {
		checks = append(checks, ScanCheckResponse{
			ID:         check.ID(),
			Type:       string(check.CheckType()),
			Status:     scanStatusToString(check.Status()),
			StartedAt:  timePtr(check.StartedAt()),
			FinishedAt: timePtr(check.FinishedAt()),
		})
	}

	return ScanResponse{
		ID:      scan.ID(),
		OwnerID: scan.OwnerID(),
		Status:  scanStatusToString(scan.Status()),
		APIType: apiTypeToString(scan.ApiType()),
		ArtifactMetadata: ArtifactMetadataResponse{
			FileName:         metadata.FileName(),
			TargetLanguage:   metadata.TargetLang(),
			SchemaVersion:    metadata.SchemaVersion(),
			SLOC:             metadata.SLOC(),
			UncompressedSize: metadata.UncompressedSize(),
			CompressedSize:   metadata.CompressedSize(),
		},
		Labels:     labels,
		ScanChecks: checks,
		CreatedAt:  scan.CreatedAt(),
		UpdatedAt:  scan.UpdatedAt(),
	}
}

func scanStatusToString(status domain.ScanStatus) string {
	switch status {
	case domain.ScanStatusPending:
		return "PENDING"
	case domain.ScanStatusRunning:
		return "RUNNING"
	case domain.ScanStatusCompleted:
		return "COMPLETED"
	case domain.ScanStatusFailed:
		return "FAILED"
	case domain.ScanStatusCancelled:
		return "CANCELLED"
	default:
		return fmt.Sprintf("UNKNOWN_%d", status)
	}
}

func apiTypeToString(apiType domain.ApiType) string {
	switch apiType {
	case domain.ApiTypeHTTP:
		return "HTTP"
	case domain.ApiTypeGRPC:
		return "GRPC"
	default:
		return fmt.Sprintf("UNKNOWN_%d", apiType)
	}
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
