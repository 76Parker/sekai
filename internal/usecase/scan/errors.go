package scan

import "errors"

var (
	ErrInvalidArtifact   = errors.New("invalid artifact")
	ErrScanQuotaExceeded = errors.New("scan quota exceeded")
)
