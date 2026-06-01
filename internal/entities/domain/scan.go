package domain

import (
	"time"
)

type ScanStatus int
type ApiType int

const (
	ScanStatusPending   ScanStatus = iota // 0 - PENDING
	ScanStatusRunning                     // 1 - RUNNING
	ScanStatusCompleted                   // 2 - COMPLETED
	ScanStatusFailed                      // 3 - FAILED
	ScanStatusCancelled                   // 4 - CANCELLED
)

const (
	ApiTypeHTTP ApiType = iota // 0 - HTTP
	ApiTypeGRPC                // 1 - GRPC
)

type Scan struct {
	id         int64
	ownerID    int64
	scanChecks []ScanCheck
	status     ScanStatus
	apiType    ApiType
	metadata   ArtifactMetadata

	createdAt time.Time
	updatedAt time.Time

	labels []Label
}

func (s Scan) ID() int64 {
	return s.id
}

func (s Scan) OwnerID() int64 {
	return s.ownerID
}

func (s Scan) ScanChecks() []ScanCheck {
	return append([]ScanCheck(nil), s.scanChecks...)
}

func (s Scan) Status() ScanStatus {
	return s.status
}

func (s Scan) ApiType() ApiType {
	return s.apiType
}

func (s Scan) Metadata() ArtifactMetadata {
	return s.metadata
}

func (s Scan) CreatedAt() time.Time {
	return s.createdAt
}

func (s Scan) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s Scan) Labels() []Label {
	return append([]Label(nil), s.labels...)
}

func NewScan(
	ownerID int64,
	apiType ApiType,
	checks []ScanCheck,
	labels []Label,
	metadata ArtifactMetadata,
) (Scan, error) {

	if ownerID <= 0 {
		return Scan{}, NewError("Scan", "ownerID must be greater than 0")
	}

	if apiType != ApiTypeHTTP && apiType != ApiTypeGRPC {
		return Scan{}, NewError("Scan", "unknown apiType")
	}

	if len(labels) > 3 {
		return Scan{}, NewError("Scan", "labels count must not exceed 3")
	}

	if len(checks) == 0 {
		return Scan{}, NewError("Scan", "scan checks must not be empty")
	}
	return Scan{
		ownerID:    ownerID,
		apiType:    apiType,
		status:     ScanStatusPending,
		scanChecks: checks,
		labels:     labels,
		metadata:   metadata,
	}, nil
}

func ReconstituteScan(id int64, ownerID int64, status ScanStatus, apiType ApiType, createdAt, updatedAt time.Time, labels []Label, checks []ScanCheck, metadata ArtifactMetadata) Scan {
	return Scan{
		id:         id,
		ownerID:    ownerID,
		status:     status,
		apiType:    apiType,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
		labels:     labels,
		scanChecks: checks,
		metadata:   metadata,
	}
}
