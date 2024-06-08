package nodes

import (
	"context"

	"github.com/warehouse/ai-service/internal/repository/models"
	"github.com/warehouse/ai-service/internal/repository/operations/transactions"
)

type Repository interface {
	GetByIds(ctx context.Context, tx transactions.Transaction, ids []string) ([]models.Node, error)
	GetById(ctx context.Context, tx transactions.Transaction, id string) (models.Node, error)

	Create(ctx context.Context, tx transactions.Transaction, node models.Node) (models.Node, error)
}
