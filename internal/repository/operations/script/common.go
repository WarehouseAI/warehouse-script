package script

import (
	"context"
	"fmt"

	"github.com/warehouse/ai-service/internal/pkg/errors/repository_errors"
	"github.com/warehouse/ai-service/internal/repository/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func (r *repositoryPG) getScriptByCondition(
	ctx context.Context,
	executor sqlx.ExtContext,
	condition string,
	params ...interface{},
) ([]models.Script, error) {
	baseQuery := `
    SELECT n.id, n.name, n.workflow, n.author, n.warehouse_api_key
  `
	query := fmt.Sprintf("%s %s", baseQuery, condition)

	var list []models.Script
	err := sqlx.SelectContext(ctx, executor, &list, query, params...)
	if err != nil {
		return nil, r.log.ErrorRepo(err, repository_errors.PostgresqlGetRaw, query)
	}

	return list, nil
}
