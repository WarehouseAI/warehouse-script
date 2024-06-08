package dependencies

import "github.com/warehouse/ai-service/internal/handler/http"

func (d *dependencies) ScriptHandler() http.Handler {
	if d.scriptHandler == nil {
		d.scriptHandler = http.NewScriptHandler(
			d.cfg.Server,
			d.cfg.Timeouts,
			d.ScriptService(),
			d.TimeAdapter(),
			d.WarehouseJsonRequestHandler(),
			d.HandlerMiddleware(),
		)
	}

	return d.scriptHandler
}

func (d *dependencies) NodeHandler() http.Handler {
	if d.nodeHandler == nil {
		d.nodeHandler = http.NewNodeHandler(
			d.cfg.Server,
			d.cfg.Timeouts,
			d.NodeService(),
			d.TimeAdapter(),
			d.WarehouseJsonRequestHandler(),
			d.HandlerMiddleware(),
		)
	}

	return d.nodeHandler
}
