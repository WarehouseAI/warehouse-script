package script

import (
	"context"
	"fmt"

	"github.com/warehouse/ai-service/internal/db"
	"github.com/warehouse/ai-service/internal/pkg/errors/repository_errors"
	"github.com/warehouse/ai-service/internal/pkg/logger"
	"github.com/warehouse/ai-service/internal/repository/models"
	"github.com/warehouse/ai-service/internal/repository/operations/transactions"
)

type repositoryPG struct {
	log logger.Logger
	pg  *db.PostgresClient
}

func NewPGRepository(log logger.Logger, client *db.PostgresClient) Repository {
	return &repositoryPG{
		pg:  client,
		log: log.Named("pg_scripts"),
	}
}

func (r *repositoryPG) GetById(ctx context.Context, tx transactions.Transaction, id string) (models.Script, error) {
	cond := `where s.id = $1`
	list, err := r.getScriptByCondition(ctx, tx.Txm(), cond, id)
	if err != nil {
		return models.Script{}, err
	}

	if len(list) != 0 {
		return list[0], nil
	} else {
		return models.Script{}, fmt.Errorf("script with provided id not found")
	}
}

func (r *repositoryPG) Create(ctx context.Context, tx transactions.Transaction, script models.Script) (models.Script, error) {
	query := `
    INSERT INTO script (name, workflow, body_presets, header_presets, author, warehouse_api_key)
    VALUES(:name, :workflow, :body_presets, :header_presets, :author, :warehouse_api_key)
  `

	res, err := tx.Txm().NamedExecContext(ctx, query, script)
	if err != nil {
		return models.Script{}, r.log.ErrorRepo(err, repository_errors.PostgresqlExecRaw, query)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return models.Script{}, r.log.ErrorRepo(err, repository_errors.PostgresqlRowsAffectedRaw, query)
	}

	if rowsAffected != 1 {
		return models.Script{}, r.log.ErrorRepo(err, repository_errors.PostgresqlRowsAffectedRaw, query)
	}

	return script, nil
}
