package node

import (
	"encoding/json"
	"fmt"

	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/pkg/errors"
)

func (s *service) validateHeader(headers map[string]interface{}) (map[string]domain.Header, *errors.Error) {
	h := make(map[string]domain.Header)
	for key, value := range headers {
		mapData, err := json.Marshal(value)
		if err != nil {
			return nil, errors.WD(errors.ParseError, err)
		}

		var header domain.Header
		if err := json.Unmarshal(mapData, &header); err != nil {
			return nil, errors.WD(errors.ParseError, err)
		}

		h[key] = header
	}

	for key, value := range h {
		if value.Type == domain.ConstHeaderType && len(value.Values) != 1 {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("header %s: with \"consts\" type there is should be one element in values array", key))
		}

		if value.Type == domain.SelectHeaderType && len(value.Values) > 2 {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("header %s: values length should be greater than 1, with type \"select\" or use \"const\" type", key))
		}

		if value.Type == domain.PromptHeaderType && len(value.Values) != 1 {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("header %s: with \"prompt\" type there is should be one element in values array", key))
		}
	}

	return h, nil
}

func (s *service) validateBody(body map[string]interface{}) (map[string]domain.BodyField, *errors.Error) {
	bodyFields := make(map[string]domain.BodyField)
	for key, value := range body {
		mapData, err := json.Marshal(value)
		if err != nil {
			return nil, errors.WD(errors.ParseError, err)
		}

		var field domain.BodyField
		if err := json.Unmarshal(mapData, &field); err != nil {
			return nil, errors.WD(errors.ParseError, fmt.Errorf("can't unmarshal field %s is inconsistent to field typing standart", key))
		}

		bodyFields[key] = field
	}

	for key, value := range bodyFields {
		if value.Type == domain.ObjectFieldType && len(value.Values) == 1 {
			switch v := value.Values[0].(type) {
			case map[string]interface{}:
				if _, e := s.validateBody(v); e != nil {
					return nil, e
				}
			default:
				return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: value is not JSON-object", key))
			}
		} else {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: only one JSON-object should be provided in values", key))
		}

		if value.Type == domain.ConstFieldType && len(value.Values) == 1 {
			if _, ok := value.Values[0].(string); !ok {
				return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: consts values should be string type", key))
			}
		} else {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: with \"consts\" type there is should be one element in values array", key))
		}

		if value.Type == domain.SelectFieldType && len(value.Values) >= 2 {
			for _, val := range value.Values {
				if _, ok := val.(string); !ok {
					return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: values in select type should be strings", key))
				}
			}
		} else {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: values length should be greater than 1, with type \"select\" or use \"const\" type", key))
		}

		if value.Type == domain.PromptFieldType && len(value.Values) == 1 {
			if _, ok := value.Values[0].(string); !ok {
				return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: prompt values should be string type", key))
			}
		} else {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: with \"propmt\" type there is should be one your custom value in values array", key))
		}

		if value.Type == domain.DataFieldType && len(value.Values) != 0 {
			return nil, errors.WD(errors.ValidationFailed, fmt.Errorf("field %s: with \"data\" type there is no values in array", key))
		}
	}

	return bodyFields, nil
}
