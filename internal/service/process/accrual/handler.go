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

	err := data.Default.Repository.SelectAccrualByIDorder(data.Default.Ctx.Request.Context(), &data.Accrual.Accrual)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	if data.Accrual.Accrual.ID != 0 {
		if data.Accrual.Accrual.IDUser != data.User.User.ID {
			data.Default.ResponseError = func() {
				data.Default.Ctx.Status(409)
			}

			return []pipeline.Message{data}, errors.New("invalid accrual upload another uers")
		} else {
			data.Default.ResponseError = func() {
				data.Default.Ctx.Status(http.StatusOK)
			}

			return []pipeline.Message{data}, errors.New("invalid accrual id order is exists")
		}

	}

	fmt.Println(data.User.User)

	return []pipeline.Message{data}, nil
}
