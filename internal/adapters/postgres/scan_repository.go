package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"sekai/internal/contracts/metadataschema"
	"sekai/internal/entities/domain"
	scanusecase "sekai/internal/usecase/scan"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScanRepository struct {
	db *pgxpool.Pool
}

func NewScanRepository(db *pgxpool.Pool) *ScanRepository {
	return &ScanRepository{db: db}
}

func (r *ScanRepository) Create(ctx context.Context, scan domain.Scan) (domain.Scan, error) {
	var created domain.Scan

	err := NewTxManager(r.db).WithinTx(ctx, func(ctx context.Context) error {
		if err := r.decrementScanQuota(ctx, scan.OwnerID()); err != nil {
			return err
		}

		var err error
		created, err = r.create(ctx, scan)
		return err
	})
	if err != nil {
		return domain.Scan{}, err
	}

	return created, nil
}

func (r *ScanRepository) ListByOwnerID(ctx context.Context, ownerID int64) ([]domain.Scan, error) {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	SELECT id, owner_id, status, api_type, artifact_schema_version, artifact_metadata, created_at, updated_at
	FROM sekai.scans
	WHERE owner_id = $1
	ORDER BY created_at DESC, id DESC;
	`

	rows, err := db.Query(ctx, sqlQuery, ownerID)
	if err != nil {
		return nil, err
	}

	type scanRecord struct {
		id                    int64
		ownerID               int64
		status                int16
		apiType               int16
		artifactSchemaVersion int
		artifactMetadata      []byte
		createdAt             time.Time
		updatedAt             time.Time
	}

	records := make([]scanRecord, 0)
	for rows.Next() {
		var record scanRecord

		if err := rows.Scan(
			&record.id,
			&record.ownerID,
			&record.status,
			&record.apiType,
			&record.artifactSchemaVersion,
			&record.artifactMetadata,
			&record.createdAt,
			&record.updatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	scans := make([]domain.Scan, 0, len(records))
	for _, record := range records {
		metadata, err := parseArtifactMetadata(record.artifactSchemaVersion, record.artifactMetadata)
		if err != nil {
			return nil, err
		}

		labels, err := r.listScanLabels(ctx, record.id)
		if err != nil {
			return nil, err
		}

		checks, err := r.listScanChecks(ctx, record.id)
		if err != nil {
			return nil, err
		}

		scans = append(scans, domain.ReconstituteScan(
			record.id,
			record.ownerID,
			domain.ScanStatus(record.status),
			domain.ApiType(record.apiType),
			record.createdAt,
			record.updatedAt,
			labels,
			checks,
			metadata,
		))
	}

	return scans, nil
}

func (r *ScanRepository) DeleteByID(ctx context.Context, id int64) error {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	DELETE FROM sekai.scans
	WHERE id = $1;
	`

	_, err := db.Exec(ctx, sqlQuery, id)
	return err
}

func (r *ScanRepository) decrementScanQuota(ctx context.Context, ownerID int64) error {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	UPDATE sekai.users
	SET scan_quota = scan_quota - 1
	WHERE id = $1
	  AND scan_quota > 0;
	`

	tag, err := db.Exec(ctx, sqlQuery, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return scanusecase.ErrScanQuotaExceeded
	}

	return nil
}

func (r *ScanRepository) create(ctx context.Context, scan domain.Scan) (domain.Scan, error) {
	db := ExecutorFromContext(ctx, r.db)

	var id int64
	now := time.Now().UTC()

	sqlQuery := `
	INSERT INTO sekai.scans (owner_id, status, api_type, artifact_schema_version, artifact_metadata, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id;
	`

	metadata := scan.Metadata()
	metadataPayload, err := json.Marshal(metadataschema.ArtifactMetadataV1{
		FileName:         metadata.FileName(),
		TargetLanguage:   metadata.TargetLang(),
		SchemaVersion:    metadata.SchemaVersion(),
		Sloc:             metadata.SLOC(),
		UncompressedSize: int(metadata.UncompressedSize()),
		CompressedSize:   int(metadata.CompressedSize()),
	})
	if err != nil {
		return domain.Scan{}, err
	}

	err = db.QueryRow(
		ctx,
		sqlQuery,
		scan.OwnerID(),
		int16(scan.Status()),
		int16(scan.ApiType()),
		metadata.SchemaVersion(),
		string(metadataPayload),
		now,
		now,
	).Scan(&id)
	if err != nil {
		return domain.Scan{}, err
	}

	createdChecks, err := r.createScanChecks(ctx, id, scan.ScanChecks())
	if err != nil {
		return domain.Scan{}, err
	}

	createdLabels, err := r.createScanLabels(ctx, id, scan.Labels())
	if err != nil {
		return domain.Scan{}, err
	}

	scan = domain.ReconstituteScan(
		id,
		scan.OwnerID(),
		domain.ScanStatus(scan.Status()),
		domain.ApiType(scan.ApiType()),
		now,
		now,
		createdLabels,
		createdChecks,
		scan.Metadata(),
	)

	return scan, nil
}

func (r *ScanRepository) createScanChecks(ctx context.Context, scanID int64, checks []domain.ScanCheck) ([]domain.ScanCheck, error) {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	INSERT INTO sekai.scan_checks (scan_id, check_type, status)
	VALUES ($1, $2, $3)
	RETURNING id;
	`

	createdChecks := make([]domain.ScanCheck, 0, len(checks))
	for _, check := range checks {
		var id int64
		if err := db.QueryRow(ctx, sqlQuery, scanID, string(check.CheckType()), int16(check.Status())).Scan(&id); err != nil {
			return nil, err
		}
		createdChecks = append(createdChecks, domain.ReconstituteScanCheck(
			id,
			scanID,
			check.CheckType(),
			check.Status(),
			time.Time{},
			time.Time{},
		))
	}

	return createdChecks, nil
}

func (r *ScanRepository) listScanLabels(ctx context.Context, scanID int64) ([]domain.Label, error) {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	SELECT id, scan_id, name
	FROM sekai.scan_labels
	WHERE scan_id = $1
	ORDER BY id ASC;
	`

	rows, err := db.Query(ctx, sqlQuery, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels := make([]domain.Label, 0)
	for rows.Next() {
		var (
			id     int64
			scanID int64
			name   string
		)
		if err := rows.Scan(&id, &scanID, &name); err != nil {
			return nil, err
		}
		labels = append(labels, domain.ReconstituteLabel(id, scanID, name))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return labels, nil
}

func (r *ScanRepository) listScanChecks(ctx context.Context, scanID int64) ([]domain.ScanCheck, error) {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	SELECT id, scan_id, check_type, status, started_at, finished_at
	FROM sekai.scan_checks
	WHERE scan_id = $1
	ORDER BY id ASC;
	`

	rows, err := db.Query(ctx, sqlQuery, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := make([]domain.ScanCheck, 0)
	for rows.Next() {
		var (
			id         int64
			scanID     int64
			checkType  string
			status     int16
			startedAt  pgtype.Timestamptz
			finishedAt pgtype.Timestamptz
		)
		if err := rows.Scan(&id, &scanID, &checkType, &status, &startedAt, &finishedAt); err != nil {
			return nil, err
		}
		checks = append(checks, domain.ReconstituteScanCheck(
			id,
			scanID,
			domain.ScanCheckType(checkType),
			domain.ScanStatus(status),
			timeFromNull(startedAt),
			timeFromNull(finishedAt),
		))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return checks, nil
}

func parseArtifactMetadata(schemaVersion int, payload []byte) (domain.ArtifactMetadata, error) {
	switch schemaVersion {
	case 1:
		var metadata metadataschema.ArtifactMetadataV1
		if err := json.Unmarshal(payload, &metadata); err != nil {
			return domain.ArtifactMetadata{}, err
		}
		if metadata.SchemaVersion == 0 {
			metadata.SchemaVersion = schemaVersion
		}

		return domain.ReconstituteArtifactMetadata(
			metadata.FileName,
			metadata.TargetLanguage,
			metadata.SchemaVersion,
			metadata.Sloc,
			int64(metadata.UncompressedSize),
			int64(metadata.CompressedSize),
		), nil
	default:
		return domain.ArtifactMetadata{}, fmt.Errorf("unsupported artifact metadata schema version: %d", schemaVersion)
	}
}

func timeFromNull(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func (r *ScanRepository) createScanLabels(ctx context.Context, scanID int64, labels []domain.Label) ([]domain.Label, error) {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	INSERT INTO sekai.scan_labels (scan_id, name)
	VALUES ($1, $2)
	RETURNING id;
	`

	createdLabels := make([]domain.Label, 0, len(labels))
	for _, label := range labels {
		var id int64
		if err := db.QueryRow(ctx, sqlQuery, scanID, label.Name()).Scan(&id); err != nil {
			return nil, err
		}
		createdLabels = append(createdLabels, domain.ReconstituteLabel(id, scanID, label.Name()))
	}

	return createdLabels, nil
}
