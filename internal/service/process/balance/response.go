package balance

import (
	"encoding/json"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type ResponseBalance struct{}

// Обрабатываем поступивший
func (m ResponseBalance) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	dataResponse, err := json.Marshal(data.Balance.ResponseBalance)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, nil
	}

	data.Default.Response = func() {
		data.Default.Ctx.JSON(http.StatusOK, string(dataResponse))
	}

	return []pipeline.Message{data}, nil
}
