package nodes

import (
	"context"
	"fmt"

	"github.com/warehouse/ai-service/internal/pkg/errors/repository_errors"
	"github.com/warehouse/ai-service/internal/repository/models"

	"github.com/jmoiron/sqlx"
)

func (r *repositoryPG) getNodeByCondition(
	ctx context.Context,
	executor sqlx.ExtContext,
	condition string,
	params ...interface{},
) ([]models.Node, error) {
	baseQuery := `
    SELECT n.id, n.name, n.url, n.method, n.headers, n.body, n.api_key
    FROM nodes as n
  `
	query := fmt.Sprintf("%s %s", baseQuery, condition)

	var list []models.Node
	err := sqlx.SelectContext(ctx, executor, &list, query, params...)
	if err != nil {
		return nil, r.log.ErrorRepo(err, repository_errors.PostgresqlGetRaw, query)
	}

	return list, nil
}
