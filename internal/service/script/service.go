package script

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/warehouse/ai-service/internal/config"
	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/handler/models"
	"github.com/warehouse/ai-service/internal/pkg/errors"
	"github.com/warehouse/ai-service/internal/pkg/logger"
	nodesRepo "github.com/warehouse/ai-service/internal/repository/operations/nodes"
	scriptRepo "github.com/warehouse/ai-service/internal/repository/operations/script"
	"github.com/warehouse/ai-service/internal/repository/operations/transactions"
)

type (
	Service interface {
		Run(ctx context.Context, request models.RunScriptRequest) (string, *errors.Error)
		Create(ctx context.Context, acc *domain.Account, request models.CreateScriptRequest) (domain.Script, *errors.Error)
	}

	service struct {
		cfg config.Config
		log logger.Logger

		txRepo     transactions.Repository
		nodesRepo  nodesRepo.Repository
		scriptRepo scriptRepo.Repository
	}
)

func NewService(
	cfg config.Config,
	log logger.Logger,
	txRepo transactions.Repository,
	nodesRepo nodesRepo.Repository,
	scriptRepo scriptRepo.Repository,
) Service {
	return &service{
		cfg:        cfg,
		log:        log,
		txRepo:     txRepo,
		nodesRepo:  nodesRepo,
		scriptRepo: scriptRepo,
	}
}

// 1. Валидация
// 2. Генерируем заполненный JSON, который потом будет передаваться в апишку
// 2. Сохранение
func (s *service) Create(ctx context.Context, acc *domain.Account, request models.CreateScriptRequest) (domain.Script, *errors.Error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return domain.Script{}, s.log.ServiceTxError(err)
	}
	defer tx.Rollback()

	workflowMap := make(map[int]map[int][]string)
	for key, value := range request.Workflow {
		stepKey, err := strconv.Atoi(key)
		if err != nil {
			return domain.Script{}, errors.WD(errors.ParseError, err)
		}

		stepChainMap := s.parseStep(value)
		workflowMap[stepKey] = stepChainMap
	}

	usedNodes, e := s.validateWorkflow(ctx, tx, request.Workflow)
	if e != nil {
		return domain.Script{}, e
	}

	if e := s.validateBodyPresets(usedNodes, request.BodyPresets); e != nil {
		return domain.Script{}, e
	}

	if e := s.validateHeaderPresets(usedNodes, request.HeaderPresets); e != nil {
		return domain.Script{}, e
	}

	// TODO: добавить айди автора
	script := domain.Script{
		Name:            request.Name,
		Workflow:        workflowMap,
		BodyPresets:     request.BodyPresets,
		HeaderPresets:   request.HeaderPresets,
		AuthorId:        acc.Id,
		WarehouseApiKey: "test_key",
	}

	modelScript, err := script.ToModel()
	if err != nil {
		return domain.Script{}, errors.WD(errors.ParseError, err)
	}

	createdScript, err := s.scriptRepo.Create(ctx, tx, modelScript)
	if err != nil {
		return domain.Script{}, errors.DatabaseError(err)
	}

	script.Id = createdScript.Id.String()

	if err := tx.Commit(); err != nil {
		return domain.Script{}, s.log.ServiceTxError(err)
	}

	return script, nil
}

func (s *service) Run(ctx context.Context, request models.RunScriptRequest) (string, *errors.Error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return "", s.log.ServiceTxError(err)
	}
	defer tx.Rollback()

	res, err := s.scriptRepo.GetById(ctx, tx, request.Id)
	if err != nil {
		return "", errors.DatabaseError(err)
	}
	script, err := domain.Script{}.FromModel(res)
	if err != nil {
		return "", errors.WD(errors.ParseError, err)
	}

	scriptMap, e := s.fillScriptMap(ctx, tx, script.Workflow)
	if e != nil {
		return "", e
	}

	stepCtx := request.EnterData
	for i := 1; i < len(scriptMap); i++ {
		step, stepOk := scriptMap[i]

		if stepOk {
			var stepWg sync.WaitGroup
			stepCh := make(chan domain.ChainResult, len(step))

			for j := 0; j < len(step); j++ {
				chain, chainOk := step[j]

				if chainOk {
					stepWg.Add(1)
					go s.chainHandler(&stepWg, stepCh, script.BodyPresets, script.HeaderPresets, chain, stepCtx)
				}
			}

			stepWg.Wait()

			// читаем данные с канала и объединяем в общий контект для следующего шага
			newContext := []string{}
			for res := range stepCh {
				if res.Error != nil {
					return "", errors.ExecError(res.Error)
				}

				newContext = append(newContext, res.Response)
			}

			stepCtx = strings.Join(newContext, ". ")
		}
	}

	if err := tx.Commit(); err != nil {
		return "", s.log.ServiceTxError(err)
	}

	return stepCtx, nil
}
