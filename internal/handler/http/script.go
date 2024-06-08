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
	"github.com/warehouse/ai-service/internal/service/script"

	"github.com/gorilla/mux"
)

type (
	scriptHandler struct {
		cfg      *config.Server
		timeouts *config.Timeouts

		scriptService script.Service

		timeAdapter timeAdpt.Adapter

		reqHandler WarehouseRequestHandler
		middleware middlewares.Middleware
	}
)

func NewScriptHandler(
	cfg config.Server,
	timeouts config.Timeouts,

	scriptSvc script.Service,

	timeAdpt timeAdpt.Adapter,

	requestHandler WarehouseRequestHandler,
	middlewares middlewares.Middleware,
) Handler {
	return &scriptHandler{
		cfg:      &cfg,
		timeouts: &timeouts,

		scriptService: scriptSvc,

		timeAdapter: timeAdpt,

		reqHandler: requestHandler,
		middleware: middlewares,
	}
}

func (h *scriptHandler) Shutdown() {
}

func (h *scriptHandler) FillHandlers(router *mux.Router) {
	base := "/script"
	r := router.PathPrefix(base).Subrouter()
	h.reqHandler.HandleJsonRequestWithMiddleware(r, base, "/run", http.MethodDelete, h.runHandler, h.middleware.JwtAuthMiddleware(domain.PurposeAccess))
	h.reqHandler.HandleJsonRequestWithMiddleware(r, base, "/create", http.MethodDelete, h.createHandler, h.middleware.JwtAuthMiddleware(domain.PurposeAccess))
}

func (h *scriptHandler) runHandler(ctx context.Context, acc *domain.Account, r *http.Request) jsonResponse {
	if acc == nil {
		return whJsonErrorResponse(errors.AuthFailed)
	}

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, h.timeouts.RequestTimeout)
	defer cancel()

	var req models.RunScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return whJsonErrorResponse(errors.WD(errors.InternalError, err))
	}

	result, err := h.scriptService.Run(ctx, req)
	if err != nil {
		return whJsonErrorResponse(err)
	}

	return whJsonSuccessResponse(
		models.RunScriptResponse{
			Result: result,
		},
		http.StatusOK,
		nil,
	)
}

func (h *scriptHandler) createHandler(ctx context.Context, acc *domain.Account, r *http.Request) jsonResponse {
	if acc == nil {
		return whJsonErrorResponse(errors.AuthFailed)
	}

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, h.timeouts.RequestTimeout)
	defer cancel()

	var req models.CreateScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return whJsonErrorResponse(errors.WD(errors.InternalError, err))
	}

	createdScript, err := h.scriptService.Create(ctx, acc, req)
	if err != nil {
		return whJsonErrorResponse(err)
	}

	return whJsonSuccessResponse(
		models.CreateScriptResponse{
			Id:            createdScript.Id,
			Name:          createdScript.Name,
			BodyPresets:   createdScript.BodyPresets,
			HeaderPresets: createdScript.HeaderPresets,
		},
		http.StatusCreated,
		nil,
	)
}
