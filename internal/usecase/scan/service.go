package scan

import (
	"context"
	"fmt"
	"sekai/internal/entities/domain"

	"github.com/google/uuid"
)

type ArtifactInspector interface {
	Inspect(ctx context.Context, archivePath string, sizeBytes int64) (InspectionInfo, error)
}

type Service struct {
	inspector ArtifactInspector
	repo      Repository
	storage   Storage
	engine    Engine
}

func NewService(inspector ArtifactInspector, repo Repository, storage Storage) *Service {
	return &Service{
		inspector: inspector,
		repo:      repo,
		storage:   storage,
	}
}

func (s *Service) Create(ctx context.Context, cmd StartCommand) (domain.Scan, error) {
	info, err := s.inspector.Inspect(ctx, cmd.ArchivePath, cmd.SizeBytes)
	if err != nil {
		return domain.Scan{}, fmt.Errorf("%w: %v", ErrInvalidArtifact, err)
	}
	checks, err := createScanChecks(cmd)
	if err != nil {
		return domain.Scan{}, err
	}
	labels, err := createLabels(cmd.Labels)
	if err != nil {
		return domain.Scan{}, err
	}
	metadata, err := domain.NewArtifactMetadata(
		info.FileName,
		info.TargetLanguage,
		1,
		info.TargetLanguageSLOC,
		info.UncompressedSize,
		info.CompressedSize,
	)
	if err != nil {
		return domain.Scan{}, err
	}

	scan, err := domain.NewScan(
		cmd.UserID,
		domain.ApiType(cmd.ApiType),
		checks,
		labels,
		metadata,
	)
	if err != nil {
		return domain.Scan{}, err
	}

	createdScan, err := s.repo.Create(ctx, scan)
	if err != nil {
		return domain.Scan{}, err
	}
	key := buildArtifactStorageKey()
	if err := s.engine.StartScan(ctx, createdScan, key); err != nil {
		return domain.Scan{}, err
	}

	return createdScan, nil
}

func (s *Service) ListByUser(ctx context.Context, userID int64) ([]domain.Scan, error) {
	if userID <= 0 {
		return nil, domain.NewError("Scan", "ownerID must be greater than 0")
	}

	return s.repo.ListByOwnerID(ctx, userID)
}

func buildArtifactStorageKey() string {
	return uuid.New().String()
}

func createScanChecks(cmd StartCommand) ([]domain.ScanCheck, error) {
	checks := []domain.ScanCheck{}
	if cmd.EnableSAST {
		check, err := domain.NewScanCheck(domain.ScanCheckSAST)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, nil
}

func createLabels(labels []string) ([]domain.Label, error) {
	labelList := []domain.Label{}
	for _, label := range labels {
		l, err := domain.NewLabel(label)
		if err != nil {
			return nil, err
		}
		labelList = append(labelList, l)
	}
	return labelList, nil
}
