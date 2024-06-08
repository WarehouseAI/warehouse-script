package script

import (
	"context"

	"github.com/warehouse/ai-service/internal/repository/models"
	"github.com/warehouse/ai-service/internal/repository/operations/transactions"
)

type Repository interface {
	GetById(ctx context.Context, tx transactions.Transaction, id string) (models.Script, error)
	Create(ctx context.Context, tx transactions.Transaction, script models.Script) (models.Script, error)
}
