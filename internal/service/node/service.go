package node

import (
	"context"

	"github.com/warehouse/ai-service/internal/config"
	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/handler/models"
	"github.com/warehouse/ai-service/internal/pkg/errors"
	"github.com/warehouse/ai-service/internal/pkg/logger"
	nodesRepo "github.com/warehouse/ai-service/internal/repository/operations/nodes"
	"github.com/warehouse/ai-service/internal/repository/operations/transactions"
)

type (
	Service interface {
		Add(ctx context.Context, request models.AddNodeRequest) (domain.Node, *errors.Error)
	}

	service struct {
		cfg config.Config
		log logger.Logger

		txRepo    transactions.Repository
		nodesRepo nodesRepo.Repository
	}
)

func NewService(
	cfg config.Config,
	log logger.Logger,
	txRepo transactions.Repository,
	nodesRepo nodesRepo.Repository,
) Service {
	return &service{
		cfg:       cfg,
		log:       log,
		txRepo:    txRepo,
		nodesRepo: nodesRepo,
	}
}

func (s *service) Add(ctx context.Context, request models.AddNodeRequest) (domain.Node, *errors.Error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return domain.Node{}, s.log.ServiceTxError(err)
	}
	defer tx.Rollback()

	fields, e := s.validateBody(request.Body)
	if e != nil {
		return domain.Node{}, e
	}

	headers, e := s.validateHeader(request.Headers)
	if e != nil {
		return domain.Node{}, e
	}

	node := domain.Node{
		Name:              request.Name,
		Url:               request.Url,
		Method:            domain.HttpMethod(request.Method),
		ResponseDirection: request.ResponseDirection,
		ApiKey:            request.ApiKey,
		RequestMime:       request.RequestMime,
		ResponseMime:      request.ResponseMime,
		Headers:           headers,
		Body:              fields,
	}

	modelNode, err := node.ToModel()
	if err != nil {
		return domain.Node{}, errors.WD(errors.ParseError, err)
	}
	createdNode, err := s.nodesRepo.Create(ctx, tx, modelNode)
	if err != nil {
		return domain.Node{}, errors.DatabaseError(err)
	}

	node.Id = createdNode.Id.String()

	if err := tx.Commit(); err != nil {
		return domain.Node{}, s.log.ServiceTxError(err)
	}

	return node, nil
}
