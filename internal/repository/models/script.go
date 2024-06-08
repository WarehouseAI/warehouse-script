package models

import (
	"encoding/json"

	"github.com/warehouse/ai-service/internal/repository/types"

	"github.com/rs/xid"
)

type (
	Script struct {
		Id              xid.ID          `db:"id"`
		Name            string          `db:"name"`
		Workflow        json.RawMessage `db:"nodes"`
		BodyPresets     types.JSON      `db:"body_presets"`
		HeaderPresets   types.JSON      `db:"header_presets"`
		AuthorId        string          `db:"author"`
		WarehouseApiKey string          `db:"warehouse_api_key"`
	}
)
