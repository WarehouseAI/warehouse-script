package domain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/warehouse/ai-service/internal/repository/models"
)

type Script struct {
	Id              string
	Name            string
	Workflow        map[int]map[int][]string
	BodyPresets     map[string]map[string]interface{}
	HeaderPresets   map[string]map[string]string
	AuthorId        string
	WarehouseApiKey string
}

func parseStep(stepData []interface{}) map[int][]string {
	stepChainMap := make(map[int][]string)

	for i, v := range stepData {
		switch value := v.(type) {
		case []interface{}:
			resMap := parseStep(value)
			stepChainMap[i] = append(stepChainMap[i], resMap[0]...)
		case string:
			stepChainMap[i] = append(stepChainMap[i], value)
		default:
			fmt.Printf("Unexpected type %T\n", value)
		}
	}

	return stepChainMap
}

func (Script) FromModel(m models.Script) (Script, error) {
	var workflow map[string][]interface{}
	if err := json.Unmarshal(m.Workflow, &workflow); err != nil {
		return Script{}, err
	}

	workflowMap := make(map[int]map[int][]string)
	for key, value := range workflow {
		stepKey, err := strconv.Atoi(key)
		if err != nil {
			return Script{}, err
		}

		stepChainMap := parseStep(value)
		workflowMap[stepKey] = stepChainMap
	}

	bodyPresets := make(map[string]map[string]interface{})
	for key, value := range m.BodyPresets {
		fields, ok := value.(map[string]interface{})
		if !ok {
			return Script{}, fmt.Errorf("can't parse model body presets for node %s", key)
		}

		bodyPresets[key] = fields
	}

	headerPresets := make(map[string]map[string]string)
	for key, value := range m.HeaderPresets {
		headers, ok := value.(map[string]string)
		if !ok {
			return Script{}, fmt.Errorf("can't parse model headers presets for node %s", key)
		}

		headerPresets[key] = headers
	}

	return Script{
		Id:              m.Id.String(),
		Name:            m.Name,
		Workflow:        workflowMap,
		BodyPresets:     bodyPresets,
		HeaderPresets:   headerPresets,
		AuthorId:        m.AuthorId,
		WarehouseApiKey: m.WarehouseApiKey,
	}, nil
}

func (s Script) ToModel() (models.Script, error) {
	flatBodyPresets := make(map[string]interface{})
	for key, value := range s.BodyPresets {
		fields, err := json.Marshal(value)
		if err != nil {
			return models.Script{}, err
		}
		flatBodyPresets[key] = fields
	}

	flatHeaderPresets := make(map[string]interface{})
	for key, value := range s.HeaderPresets {
		headers, err := json.Marshal(value)
		if err != nil {
			return models.Script{}, err
		}
		flatHeaderPresets[key] = headers
	}

	workflow := make(map[string][]interface{})
	for stepNumber, stepValue := range s.Workflow {
		step := make([]interface{}, len(stepValue))

		for _, chainValue := range stepValue {
			step = append(step, chainValue)
		}

		workflow[strconv.Itoa(stepNumber)] = step
	}

	workflowRaw, err := json.Marshal(workflow)
	if err != nil {
		return models.Script{}, nil
	}

	return models.Script{
		Name:            s.Name,
		Workflow:        workflowRaw,
		BodyPresets:     flatBodyPresets,
		HeaderPresets:   flatHeaderPresets,
		AuthorId:        s.AuthorId,
		WarehouseApiKey: s.WarehouseApiKey,
	}, nil
}
