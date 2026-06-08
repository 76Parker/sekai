package temporal

import (
	"context"
	"sekai/internal/entities/domain"

	"go.temporal.io/sdk/client"
)

type WorkflowStarter struct {
	client client.Client
}

func (ws *WorkflowStarter) StartScan(ctx context.Context, scan domain.Scan, artifactKey string) error {
	return nil
}
