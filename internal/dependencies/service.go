package dependencies

import (
	"github.com/warehouse/ai-service/internal/service/node"
	"github.com/warehouse/ai-service/internal/service/script"
)

func (d *dependencies) ScriptService() script.Service {
	if d.scriptService == nil {
		d.scriptService = script.NewService(
			*d.cfg,
			d.log,
			d.pgxTransactionRepo,
			d.nodesRepo,
			d.scriptRepo,
		)
	}

	return d.scriptService
}

func (d *dependencies) NodeService() node.Service {
	if d.nodeService == nil {
		d.nodeService = node.NewService(
			*d.cfg,
			d.log,
			d.pgxTransactionRepo,
			d.nodesRepo,
		)
	}

	return d.nodeService
}
