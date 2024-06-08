package models

import "github.com/warehouse/ai-service/internal/domain"

type (
	AddNodeRequest struct {
		Name              string                 `json:"name"`
		Body              map[string]interface{} `json:"body"`
		RequestMime       string                 `json:"request_mime"`
		ResponseMime      string                 `json:"response_mime"`
		Url               string                 `json:"url"`
		Method            string                 `json:"method"`
		Headers           map[string]interface{} `json:"headers"`
		ResponseDirection string                 `json:"response_direction"`
		ApiKey            string                 `json:"api_key"`
	}

	AddNodeResponse struct {
		Id     string                      `json:"id"`
		Body   map[string]domain.BodyField `json:"body"`
		Header map[string]domain.Header    `json:"header"`
	}
)
