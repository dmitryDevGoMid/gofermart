package accrual

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/luna"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/opentracing/opentracing-go"
)

type AccrualCheckAlgoritmLuna struct{}

// Обрабатываем поступивший
func (m AccrualCheckAlgoritmLuna) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	span, _ := opentracing.StartSpanFromContext(ctx, "Service.Process.HandlerAccrual")
	defer span.Finish()

	data := result.(*service.Data)

	accrual := &repository.Accrual{}

	accrual.IDorder = string(data.Default.Body)

	err := luna.Validate(accrual.IDorder)

	//Проверяем номер на валидность
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(422)
		}
		return []pipeline.Message{data}, errors.New("invalid check number")
	}

	data.Accrual.Accrual = *accrual

	return []pipeline.Message{data}, nil
}
