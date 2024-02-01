package accrual

import (
	"context"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type ResponseAccrual struct{}

// Обрабатываем поступивший
func (m ResponseAccrual) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	span, _ := data.Default.Tracing.Tracing(ctx, "Service.Process.ResponseAccrual")
	if span != nil {
		defer span.Finish()
	}

	data.Default.Response = func() {
		data.Default.Ctx.Status(202)
	}

	return []pipeline.Message{data}, nil
}
