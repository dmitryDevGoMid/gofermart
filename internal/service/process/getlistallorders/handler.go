package getlistallorders

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerGetListAllOrdersByAccrual struct{}

// Обрабатываем поступившие данные
func (m HandlerGetListAllOrdersByAccrual) Process(result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerGetListAllOrders")

	data := result.(*service.Data)

	listOrders, err := data.Default.Repository.SelectAccrualByUser(data.Default.Ctx.Request.Context(), &data.User.User)

	fmt.Println(listOrders)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	if len(listOrders) == 0 {
		data.Default.Response = func() {
			data.Default.Ctx.Status(204)
		}
		return []pipeline.Message{data}, errors.New("empty data response list accrual")
	}

	data.Accrual.AccrualList = &listOrders

	return []pipeline.Message{data}, nil
}
