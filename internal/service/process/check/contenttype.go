package check

import (
	"context"
	"errors"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type CheckContentTypeOrders struct{}

// Обрабатываем поступивший
func (m CheckContentTypeWithdraw) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	if data.Default.Ctx.Request.Header.Get("Content-Type") != "application/json" {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, errors.New("Content-Type is Bad")
	}

	return []pipeline.Message{data}, nil
}

type CheckContentTypeWithdraw struct{}

// Обрабатываем поступивший
func (m CheckContentTypeOrders) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	if data.Default.Ctx.Request.Header.Get("Content-Type") != "text/plain" {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, errors.New("Content-Type is Bad")
	}

	return []pipeline.Message{data}, nil
}
