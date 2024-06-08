package dependencies

import (
	"github.com/warehouse/ai-service/internal/repository/operations/nodes"
	"github.com/warehouse/ai-service/internal/repository/operations/script"
	"github.com/warehouse/ai-service/internal/repository/operations/transactions"
)

func (d *dependencies) PgxTransactionRepo() transactions.Repository {
	if d.pgxTransactionRepo == nil {
		d.pgxTransactionRepo = transactions.NewPgxRepository(d.PostgresClient())
	}
	return d.pgxTransactionRepo
}

func (d *dependencies) ScriptRepo() script.Repository {
	if d.scriptRepo == nil {
		d.scriptRepo = script.NewPGRepository(d.log, d.PostgresClient())
	}

	return d.scriptRepo
}

func (d *dependencies) NodesRepo() nodes.Repository {
	if d.nodesRepo == nil {
		d.nodesRepo = nodes.NewPGRepository(d.log, d.PostgresClient())
	}

	return d.nodesRepo
}
