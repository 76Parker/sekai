package domain

import "time"

type ScanCheckType string

var (
	ScanCheckSAST    ScanCheckType = "SAST"
	ScanCheckDAST    ScanCheckType = "DAST"
	ScanCheckSCA     ScanCheckType = "SCA"
	ScanCheckSecrets ScanCheckType = "SECRETS"
)

type ScanCheck struct {
	id        int64
	scanID    int64
	checkType ScanCheckType
	status    ScanStatus

	startedAt  time.Time
	finishedAt time.Time
}

func (c ScanCheck) CheckType() ScanCheckType {
	return c.checkType
}

func (c ScanCheck) ID() int64 {
	return c.id
}

func (c ScanCheck) ScanID() int64 {
	return c.scanID
}

func (c ScanCheck) Status() ScanStatus {
	return c.status
}

func (c ScanCheck) StartedAt() time.Time {
	return c.startedAt
}

func (c ScanCheck) FinishedAt() time.Time {
	return c.finishedAt
}

func NewScanCheck(
	checkType ScanCheckType,
) (ScanCheck, error) {
	if checkType != ScanCheckSAST && checkType != ScanCheckDAST && checkType != ScanCheckSCA && checkType != ScanCheckSecrets {
		return ScanCheck{}, NewError("ScanCheck", "unknown checkType")
	}

	return ScanCheck{
		checkType: checkType,
		status:    ScanStatusPending,
	}, nil
}

func ReconstituteScanCheck(
	id int64,
	scanID int64,
	checkType ScanCheckType,
	status ScanStatus,
	startedAt time.Time,
	finishedAt time.Time,
) ScanCheck {
	return ScanCheck{
		id:         id,
		scanID:     scanID,
		checkType:  checkType,
		status:     status,
		startedAt:  startedAt,
		finishedAt: finishedAt,
	}
}
