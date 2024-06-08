package script

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/warehouse/ai-service/internal/domain"
)

type (
	nodeHandler struct {
		requestTimeout time.Duration
	}
)

func newNodeHandler(
	requestTimeout time.Duration,
) nodeHandler {
	return nodeHandler{
		requestTimeout: requestTimeout,
	}
}

// В редакторе скрипта, мы уже вносим необходимые настройки, поэтому мы сохраняем json запроса, в который нужно лишь подставить пользовательский запрос/промпт
// Для этого запроса надо сделать путь как с ответом
// TODO: Подумать над контекстами (10 секунд 100% мало и надо сделать отдельный чисто для запросов к иишкам)
func (s *nodeHandler) makeHTTPRequest(
	node domain.Node,
	headers map[string]string,
	request []byte,
) ([]byte, error) {
	var buffer bytes.Buffer
	httpClient := http.Client{}

	url, err := url.Parse(node.Url)
	if err != nil {
		return nil, err
	}

	if err := json.Compact(&buffer, request); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(string(node.Method), url.String(), &buffer)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
