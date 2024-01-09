package getlistallwithdrawals

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerGetListAllOrdersByWithDraw struct{}

// Обрабатываем поступившие данные
func (m HandlerGetListAllOrdersByWithDraw) Process(result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerGetListAllOrdersByWithDraw")

	data := result.(*service.Data)

	listWithdrawals, err := data.Default.Repository.SelectWithdrawByUsers(data.Default.Ctx.Request.Context(), &data.User.User)

	fmt.Println(listWithdrawals)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	if len(listWithdrawals) == 0 {
		data.Default.Response = func() {
			data.Default.Ctx.Status(204)
		}
		return []pipeline.Message{data}, errors.New("empty data response list drawals")
	}

	data.Withdraw.WithdrawList = &listWithdrawals

	return []pipeline.Message{data}, nil
}
