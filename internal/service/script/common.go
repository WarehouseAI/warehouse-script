package script

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/pkg/errors"
	"github.com/warehouse/ai-service/internal/repository/operations/transactions"

	"github.com/thedevsaddam/gojsonq/v2"
)

func (s *service) chainHandler(
	stepWg *sync.WaitGroup,
	stepCh chan domain.ChainResult,
	bodyPresets map[string]map[string]interface{},
	headerPresets map[string]map[string]string,
	chain []domain.Node,
	prompt string,
) {
	defer stepWg.Done()
	nodeHandler := newNodeHandler(s.cfg.Timeouts.RequestTimeout)
	jsonq := gojsonq.New()

	finalMime := domain.JsonContentType
	for _, node := range chain {
		requestBody, err := s.generateNodeFilledObject(node.Body, prompt, bodyPresets[node.Id])
		if err != nil {
			stepCh <- domain.ChainResult{
				Response: "",
				Mime:     "",
				Error:    err,
			}
			return
		}
		marshaledBody, err := json.Marshal(requestBody)
		if err != nil {
			stepCh <- domain.ChainResult{
				Response: "",
				Mime:     "",
				Error:    err,
			}
			return
		}

		r, err := nodeHandler.makeHTTPRequest(node, headerPresets[node.Id], marshaledBody)
		if err != nil {
			stepCh <- domain.ChainResult{
				Response: "",
				Mime:     "",
				Error:    err,
			}
			return
		}

		prompt = jsonq.FromString(string(r)).Find(node.ResponseDirection).(string)
		finalMime = node.ResponseMime
	}

	stepCh <- domain.ChainResult{
		Response: prompt,
		Mime:     finalMime,
		Error:    nil,
	}
}

func (s *service) fillScriptMap(ctx context.Context, tx transactions.Transaction, workflow map[int]map[int][]string) (map[int]map[int][]domain.Node, *errors.Error) {
	nodesIds := []string{}
	for _, step := range workflow {
		for _, id := range step {
			nodesIds = append(nodesIds, id...)
		}
	}

	nodes, err := s.nodesRepo.GetByIds(ctx, tx, nodesIds)
	if err != nil {
		return nil, errors.DatabaseError(err)
	}

	nodesMap := make(map[string]domain.Node)
	for _, node := range nodes {
		nodesMap[node.Id.String()], err = domain.Node{}.FromModel(node)
		if err != nil {
			return nil, errors.WD(errors.ParseError, err)
		}
	}

	scriptFilledMap := make(map[int]map[int][]domain.Node)
	for i, step := range workflow {
		filledStep := make(map[int][]domain.Node)
		for k, chain := range step {
			filledChain := make([]domain.Node, len(chain))

			// Проверяем, что такая нода существует в скрипте, если нет -> возвращаем ошибку сразу
			for j, nodeId := range chain {
				node, ok := nodesMap[nodeId]
				if !ok {
					return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("node with id %s not found", nodeId))
				}

				filledChain[j] = node
			}

			filledStep[k] = filledChain
		}

		scriptFilledMap[i] = filledStep
	}

	return scriptFilledMap, nil
}

// Мапа мапы потому что сначала айдишник ноды потом ключ значение параметра в боди
func (s *service) validateBodyPresets(usedNodes map[string]domain.Node, bodyPresets map[string]map[string]interface{}) *errors.Error {
	// заполняем мапу для простого получения инфы по нодам и мапу с обязательными полями для каждой ноды
	requiredNodeField := make(map[string]map[string]bool)
	for id, node := range usedNodes {
		requiredFields := make(map[string]bool)
		for key, value := range node.Body {
			if value.Required {
				requiredFields[key] = true
			}
		}
		requiredNodeField[id] = requiredFields
	}

	// проходимся по пресетам и проверяем что:
	// 1. все обязательные поля/пресеты объявлены
	// 2. данные в пресете совпадают с доступными данными, в соответствии с типом поля
	for nodeId, nodePresets := range bodyPresets {
		node := usedNodes[nodeId]
		requiredFields := requiredNodeField[node.Id]

		for fieldName, fieldValue := range nodePresets {
			field, exists := node.Body[fieldName]
			if !exists {
				return errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: not found in original node body", fieldName))
			}

			if err := field.CheckBodyField(fieldValue); err != nil {
				return errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: %s", fieldName, err.Error()))
			}

			delete(requiredFields, fieldName)
		}

		if len(requiredFields) != 0 {
			missedFields := []string{}
			for key := range requiredFields {
				missedFields = append(missedFields, key)
			}

			return errors.WD(errors.ValidationFailed, fmt.Errorf("required fields [%s] have not been declared", strings.Join(missedFields, ", ")))
		}
	}

	return nil
}

func (s *service) validateHeaderPresets(usedNodes map[string]domain.Node, headerPresets map[string]map[string]string) *errors.Error {
	requiredNodeHeaders := make(map[string]map[string]bool)
	for id, node := range usedNodes {
		requiredHeaders := make(map[string]bool)
		for key, value := range node.Headers {
			if value.Required {
				requiredHeaders[key] = true
			}
		}
		requiredNodeHeaders[id] = requiredHeaders
	}

	for nodeId, headerPresets := range headerPresets {
		node := usedNodes[nodeId]
		requiredHeaders := requiredNodeHeaders[nodeId]

		for headerName, headerValue := range headerPresets {
			header, exists := node.Headers[headerName]
			if !exists {
				return errors.WD(errors.ValidationFailed, fmt.Errorf("header %s: not found in original node headers", headerName))
			}

			if err := header.CheckHeader(headerValue); err != nil {
				return errors.WD(errors.ValidationFailed, fmt.Errorf("header %s: %s", headerName, err.Error()))
			}

			delete(requiredHeaders, headerName)
		}

		if len(requiredHeaders) != 0 {
			missedHeaders := []string{}
			for key := range requiredHeaders {
				missedHeaders = append(missedHeaders, key)
			}

			return errors.WD(errors.ValidationFailed, fmt.Errorf("required headers [%s] have not been provided", strings.Join(missedHeaders, ", ")))
		}
	}

	return nil
}

func (s *service) validateWorkflow(ctx context.Context, tx transactions.Transaction, workflow map[string][]interface{}) (map[string]domain.Node, *errors.Error) {
	usedNodes := []string{}

	for _, values := range workflow {
		for _, val := range values {
			switch v := val.(type) {
			case []string:
				usedNodes = append(usedNodes, v...)

			case string:
				usedNodes = append(usedNodes, v)
			default:
				return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("workflow should by map of only strings, where string is node id"))
			}
		}
	}

	nodes, err := s.nodesRepo.GetByIds(ctx, tx, usedNodes)
	if err != nil {
		return nil, errors.DatabaseError(err)
	}

	usedNodesMap := make(map[string]domain.Node)
	for _, node := range nodes {
		usedNodesMap[node.Id.String()], err = domain.Node{}.FromModel(node)
		if err != nil {
			return nil, errors.WD(errors.ParseError, err)
		}
	}

	return usedNodesMap, nil
}

func (s *service) parseStep(stepData []interface{}) map[int][]string {
	stepChainMap := make(map[int][]string)

	for i, v := range stepData {
		switch value := v.(type) {
		case []interface{}:
			resMap := s.parseStep(value)
			stepChainMap[i] = append(stepChainMap[i], resMap[0]...)
		case string:
			stepChainMap[i] = append(stepChainMap[i], value)
		default:
			fmt.Printf("Unexpected type %T\n", value)
		}
	}

	return stepChainMap
}

func (s *service) generateNodeFilledObject(
	fields map[string]domain.BodyField,
	data string,
	bodyPresets map[string]interface{},
) (map[string]interface{}, error) {
	generatedJson := make(map[string]interface{})

	for name, value := range fields {
		if _, ok := bodyPresets[name]; ok {
			switch value.Type {
			case domain.PromptFieldType:
				generatedJson[name] = bodyPresets[name].(string)
			case domain.ConstFieldType:
				generatedJson[name] = bodyPresets[name].(string)
			case domain.SelectFieldType:
				generatedJson[name] = bodyPresets[name].(string)
			case domain.ObjectFieldType:
				nestedFieldsTyped := make(map[string]domain.BodyField)
				nestedFields, ok := value.Values[0].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("can't parse to nested structure")
				}

				for nestedName, nestedValue := range nestedFields {
					obj, err := json.Marshal(nestedValue)
					if err != nil {
						return nil, err
					}

					var nestedValueTyped domain.BodyField
					if err := json.Unmarshal(obj, &nestedValueTyped); err != nil {
						return nil, err
					}

					nestedFieldsTyped[nestedName] = nestedValueTyped
				}

				nestedBodyPresets, ok := bodyPresets[name].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("can't parse to nested structure")
				}
				filledNestedFields, err := s.generateNodeFilledObject(nestedFieldsTyped, data, nestedBodyPresets)
				if err != nil {
					return nil, err
				}

				generatedJson[name] = filledNestedFields
			case domain.DataFieldType:
				generatedJson[name] = data
			}
		}
	}

	return generatedJson, nil
}
