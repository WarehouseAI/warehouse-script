package models

import (
	"github.com/warehouse/ai-service/internal/repository/types"

	"github.com/rs/xid"
)

type (
	Node struct {
		Id                xid.ID     `db:"id"`
		Name              string     `db:"name"`
		Url               string     `db:"url"`
		Method            string     `db:"method"`
		Headers           types.JSON `db:"headers"`
		Body              types.JSON `db:"body"`
		ResponseDirection string     `db:"response_direction"` // какое поле будет передано следующей ноде как запрос или в финальный ответ, только для json'ов
		RequestMime       string     `db:"request_mime"`       // тип body, который принимает нода
		ResponseMime      string     `db:"response_mime"`      // mime type ответа
		ApiKey            string     `db:"api_key"`
	}
)
