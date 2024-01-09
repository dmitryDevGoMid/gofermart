package withdraw

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerWithdraw struct{}

// Обрабатываем поступившие данные
func (m HandlerWithdraw) Process(result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	//Сумма списаний
	totalWithdraw, err := data.Default.Repository.SelectWithdrawByUserSum(data.Default.Ctx.Request.Context(), &data.User.User)
	fmt.Println(totalWithdraw)
	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	//Сумма начислений
	totalAccrual, err := data.Default.Repository.SelectAccrualByUserSum(data.Default.Ctx.Request.Context(), &data.User.User)
	fmt.Println(totalAccrual)
	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	//Разница
	calcTotal := totalAccrual - totalWithdraw

	fmt.Println("Списание:", data.Withdraw.Withdraw.Sum)
	//Сравниваем остаток и сумму списания по заказу
	if calcTotal < data.Withdraw.Withdraw.Sum {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(402)
		}

		return []pipeline.Message{data}, errors.New("total is smaller than withdraw")
	}

	return []pipeline.Message{data}, nil
}
