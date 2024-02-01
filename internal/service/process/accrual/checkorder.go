package accrual

import (
	"context"
	"fmt"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerAccrualCheckOrder struct{}

// Обрабатываем поступивший
func (m HandlerAccrualCheckOrder) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	span, _ := data.Default.Tracing.Tracing(ctx, "Service.Process.HandlerAccrualCheckOrder")
	if span != nil {
		defer span.Finish()
	}

	accrual := &repository.Accrual{}

	accrual.IDorder = string(data.Default.Body)

	fmt.Println(*accrual)

	return []pipeline.Message{data}, nil
}
