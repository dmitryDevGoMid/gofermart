package balance

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/opentracing/opentracing-go"
)

type ResponseBalance struct{}

// Обрабатываем поступивший
func (m ResponseBalance) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Service.Process.ResponseBalance")
	defer span.Finish()

	data := result.(*service.Data)

	dataResponse, err := json.Marshal(data.Balance.ResponseBalance)

	if err != nil {

		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, nil
	}

	data.Default.Response = func() {
		data.Default.Ctx.Data(http.StatusOK, "application/json", dataResponse)
	}

	return []pipeline.Message{data}, nil
}
