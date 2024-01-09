package balance

import (
	"encoding/json"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline2"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type ResponseBalance struct{}

// Обрабатываем поступивший
func (m ResponseBalance) Process(result pipeline2.Message) ([]pipeline2.Message, error) {
	data := result.(*service.Data)

	dataResponse, err := json.Marshal(data.Balance.ResponseBalance)

	if err != nil {
		/*data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}*/
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline2.Message{data}, nil
	}

	data.Default.Response = func() {
		data.Default.Ctx.Data(http.StatusOK, "application/json", dataResponse)
	}

	return []pipeline2.Message{data}, nil
}
