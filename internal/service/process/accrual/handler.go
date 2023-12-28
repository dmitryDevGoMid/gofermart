package accrual

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerAccrual struct{}

// Обрабатываем поступившие данные
func (m HandlerAccrual) Process(result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	err := data.Default.Repository.SelectAccrualByIdOrder(data.Default.Ctx.Request.Context(), &data.Accrual.Accrual)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	if data.Accrual.Accrual.Id != 0 {
		if data.Accrual.Accrual.IdUser != data.User.User.Id {
			data.Default.ResponseError = func() {
				data.Default.Ctx.Status(409)
			}

			return []pipeline.Message{data}, errors.New("Invalid Accrual Upload Another Users")
		} else {
			data.Default.ResponseError = func() {
				data.Default.Ctx.Status(http.StatusOK)
			}

			return []pipeline.Message{data}, errors.New("Invalid Accrual ID ORDER IS EXISTS")
		}

	}

	fmt.Println(data.User.User)

	return []pipeline.Message{data}, nil
}
