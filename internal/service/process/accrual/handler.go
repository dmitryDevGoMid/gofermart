package accrual

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerAccrual struct{}

// Обрабатываем поступившие данные
func (m HandlerAccrual) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.HandlerAccrual")
	if span != nil {
		defer span.Finish()
	}
	fmt.Println("Execute HandlerAccrual2")
	dataAccrual, err := data.Default.Repository.SelectAccrualByIDorder(ctx, &data.Accrual.Accrual)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusInternalServerError)
		}

		return []pipeline.Message{data}, err
	}

	if dataAccrual != nil {
		if dataAccrual.IDUser != data.User.User.ID {
			data.Default.Response = func() {
				data.Default.Ctx.Status(409)
			}

			return []pipeline.Message{data}, errors.New("invalid accrual upload another uers")
		} else {
			data.Default.Response = func() {
				data.Default.Ctx.Status(http.StatusOK)
			}

			return []pipeline.Message{data}, errors.New("invalid accrual id order is exists")
		}

	}

	fmt.Println(data.User.User)

	return []pipeline.Message{data}, nil
}
