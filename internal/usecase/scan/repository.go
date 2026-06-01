package scan

import (
	"context"
	"sekai/internal/entities/domain"
)

type Repository interface {
	Create(ctx context.Context, scan domain.Scan) (domain.Scan, error)
	ListByOwnerID(ctx context.Context, ownerID int64) ([]domain.Scan, error)
	DeleteByID(ctx context.Context, id int64) error
}

type Storage interface {
	Save(ctx context.Context, key string, artifact domain.Scan) error
	DeleteByKey(ctx context.Context, key string) error
}
