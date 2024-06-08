package nodes

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
		log: log.Named("pg_nodes"),
	}
}

func (r *repositoryPG) Create(ctx context.Context, tx transactions.Transaction, node models.Node) (models.Node, error) {
	query := `
    INSERT INTO nodes (name, url, api_key, method, headers)
    VALUES(:name, :url, :api_key, :method, :headers)
  `

	res, err := tx.Txm().NamedExecContext(ctx, query, node)
	if err != nil {
		return models.Node{}, r.log.ErrorRepo(err, repository_errors.PostgresqlExecRaw, query)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return models.Node{}, r.log.ErrorRepo(err, repository_errors.PostgresqlRowsAffectedRaw, query)
	}

	if rowsAffected != 1 {
		return models.Node{}, r.log.ErrorRepo(err, repository_errors.PostgresqlRowsAffectedRaw, query)
	}

	return node, nil
}

func (r *repositoryPG) GetById(ctx context.Context, tx transactions.Transaction, id string) (models.Node, error) {
	cond := `WHERE n.id = $1`
	list, err := r.getNodeByCondition(ctx, tx.Txm(), cond, id)
	if err != nil {
		return models.Node{}, err
	}

	if len(list) != 0 {
		return list[0], nil
	} else {
		return models.Node{}, fmt.Errorf("node with provided id not found")
	}
}

func (r *repositoryPG) GetByIds(ctx context.Context, tx transactions.Transaction, ids []string) ([]models.Node, error) {
	cond := `WHERE n.id = ANY($1)`
	list, err := r.getNodeByCondition(ctx, tx.Txm(), cond, ids)
	if err != nil {
		return nil, err
	}

	if len(list) != 0 {
		return list, nil
	} else {
		return nil, fmt.Errorf("nodes with provided ids not found")
	}
}
