package models

type (
	RunScriptRequest struct {
		Id        string `json:"id"`
		EnterData string `json:"enter_data"`
	}
	RunScriptResponse struct {
		Result string `json:"result"`
	}

	CreateScriptRequest struct {
		Name          string                            `json:"name"`
		Workflow      map[string][]interface{}          `json:"workflow"`
		BodyPresets   map[string]map[string]interface{} `json:"body_presets"`
		HeaderPresets map[string]map[string]string      `json:"header_presets"`
	}

	CreateScriptResponse struct {
		Id            string                            `json:"id"`
		Name          string                            `json:"name"`
		BodyPresets   map[string]map[string]interface{} `json:"body_presets"`
		HeaderPresets map[string]map[string]string      `json:"header_presets"`
	}
)
