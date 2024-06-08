package http

import (
	"context"
	"encoding/json"
	"net/http"

	timeAdpt "github.com/warehouse/ai-service/internal/adapter/time"
	"github.com/warehouse/ai-service/internal/config"
	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/handler/middlewares"
	"github.com/warehouse/ai-service/internal/handler/models"
	"github.com/warehouse/ai-service/internal/pkg/errors"
	"github.com/warehouse/ai-service/internal/service/node"

	"github.com/gorilla/mux"
)

type (
	nodeHandler struct {
		cfg      *config.Server
		timeouts *config.Timeouts

		nodeService node.Service

		timeAdapter timeAdpt.Adapter

		reqHandler WarehouseRequestHandler
		middleware middlewares.Middleware
	}
)

func NewNodeHandler(
	cfg config.Server,
	timeouts config.Timeouts,

	nodeSvc node.Service,

	timeAdpt timeAdpt.Adapter,

	requestHandler WarehouseRequestHandler,
	middlewares middlewares.Middleware,
) Handler {
	return &nodeHandler{
		cfg:      &cfg,
		timeouts: &timeouts,

		nodeService: nodeSvc,

		timeAdapter: timeAdpt,

		reqHandler: requestHandler,
		middleware: middlewares,
	}
}

func (h *nodeHandler) Shutdown() {
}

func (h *nodeHandler) FillHandlers(router *mux.Router) {
	base := "/node"
	r := router.PathPrefix(base).Subrouter()
	h.reqHandler.HandleJsonRequestWithMiddleware(r, base, "/add", http.MethodDelete, h.addHandler, h.middleware.JwtAuthMiddleware(domain.PurposeAccess))
}

func (h *nodeHandler) addHandler(ctx context.Context, acc *domain.Account, r *http.Request) jsonResponse {
	if acc == nil {
		return whJsonErrorResponse(errors.AuthFailed)
	}

	if ok := rolesPermissionsInterceptor(acc.Role, domain.RoleAdmin); !ok {
		return whJsonErrorResponse(errors.PermissionDenied)
	}

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, h.timeouts.RequestTimeout)
	defer cancel()

	var req models.AddNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return whJsonErrorResponse(errors.WD(errors.InternalError, err))
	}

	createdNode, err := h.nodeService.Add(ctx, req)
	if err != nil {
		return whJsonErrorResponse(err)
	}

	return whJsonSuccessResponse(
		models.AddNodeResponse{
			Id:     createdNode.Id,
			Body:   createdNode.Body,
			Header: createdNode.Headers,
		},
		http.StatusCreated,
		nil,
	)
}
