package domain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/warehouse/ai-service/internal/repository/models"
)

type HttpMethod string

const (
	POST  HttpMethod = http.MethodPost
	GET   HttpMethod = http.MethodGet
	PATCH HttpMethod = http.MethodPatch
)

type BodyFieldType string

const (
	SelectFieldType BodyFieldType = "select"
	ConstFieldType  BodyFieldType = "const"
	ObjectFieldType BodyFieldType = "object"
	PromptFieldType BodyFieldType = "prompt"
	DataFieldType   BodyFieldType = "data"
)

type HeaderType string

const (
	SelectHeaderType HeaderType = "select"
	ConstHeaderType  HeaderType = "const"
	PromptHeaderType HeaderType = "prompt"
)

type Node struct {
	Id                string
	Name              string
	Url               string
	Method            HttpMethod
	Headers           map[string]Header
	Body              map[string]BodyField
	ResponseDirection string
	RequestMime       string
	ResponseMime      string
	ApiKey            string
}

type BodyField struct {
	Type     BodyFieldType `json:"type"`
	Values   []interface{} `json:"values"`
	Required bool          `json:"required"`
}

type Header struct {
	Type     HeaderType `json:"type"`
	Values   []string   `json:"values"`
	Required bool       `json:"required"`
}

func (bd BodyField) CheckBodyField(value interface{}) error {
	if bd.Type == ConstFieldType {
		stringValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("should contain only string values")
		}

		if bd.Values[0] != stringValue {
			return fmt.Errorf("value non equals to const value %s", bd.Values[0])
		}
	}

	if bd.Type == SelectFieldType {
		stringValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("should contain only string values")
		}

		values := make([]string, len(bd.Values))
		for i, val := range bd.Values {
			values[i] = val.(string)
		}

		if !slices.Contains(values, stringValue) {
			return fmt.Errorf("value non equals to available data [%s]", strings.Join(values, ", "))
		}
	}

	if bd.Type == ObjectFieldType {
		nestedObject, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("should be JSON object")
		}

		nestedObjectTypings := make(map[string]BodyField)
		for key, value := range bd.Values[0].(map[string]interface{}) {
			mapData, err := json.Marshal(value)
			if err != nil {
				return err
			}

			var field BodyField
			if err := json.Unmarshal(mapData, &field); err != nil {
				return fmt.Errorf("can't unmarshal is inconsistent to field typing standart")
			}

			nestedObjectTypings[key] = field
		}

		for key, value := range nestedObjectTypings {
			if err := value.CheckBodyField(nestedObject[key]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h Header) CheckHeader(value string) error {
	if h.Type == ConstHeaderType && h.Values[0] != value {
		return fmt.Errorf("header value non equals to const value %s", h.Values[0])
	}

	if h.Type == SelectHeaderType && !slices.Contains(h.Values, value) {
		return fmt.Errorf("header value non equals to available data [%s]", strings.Join(h.Values, ", "))
	}

	return nil
}

func (n Node) ToModel() (models.Node, error) {
	body := make(map[string]interface{})
	for key, value := range n.Body {
		var val map[string]interface{}
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return models.Node{}, err
		}
		if err := json.Unmarshal(jsonValue, &val); err != nil {
			return models.Node{}, err
		}

		body[key] = val
	}

	headers := make(map[string]interface{})
	for key, value := range n.Headers {
		var val map[string]interface{}
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return models.Node{}, err
		}
		if err := json.Unmarshal(jsonValue, &val); err != nil {
			return models.Node{}, err
		}

		headers[key] = val
	}

	return models.Node{
		Name:              n.Name,
		Url:               n.Url,
		Method:            string(n.Method),
		ResponseDirection: n.ResponseDirection,
		RequestMime:       n.RequestMime,
		ResponseMime:      n.RequestMime,
		ApiKey:            n.ApiKey,
		Headers:           headers,
		Body:              body,
	}, nil
}

func (Node) FromModel(m models.Node) (Node, error) {
	bodyFields := make(map[string]BodyField)
	for key, value := range m.Body {
		mapData, err := json.Marshal(value)
		if err != nil {
			return Node{}, err
		}

		var field BodyField
		if err := json.Unmarshal(mapData, &field); err != nil {
			return Node{}, err
		}

		bodyFields[key] = field
	}

	headers := make(map[string]Header)
	for key, value := range m.Headers {
		mapData, err := json.Marshal(value)
		if err != nil {
			return Node{}, err
		}

		var header Header
		if err := json.Unmarshal(mapData, &header); err != nil {
			return Node{}, err
		}

		headers[key] = header
	}

	return Node{
		Id:                m.Id.String(),
		Name:              m.Name,
		Url:               m.Url,
		Method:            HttpMethod(m.Method),
		Headers:           headers,
		Body:              bodyFields,
		ResponseDirection: m.ResponseDirection,
		RequestMime:       m.RequestMime,
		ResponseMime:      m.ResponseMime,
		ApiKey:            m.ApiKey,
	}, nil
}
