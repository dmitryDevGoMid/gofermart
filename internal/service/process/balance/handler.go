package balance

import (
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerBalance struct{}

// Обрабатываем поступившие данные
func (m HandlerBalance) Process(result pipeline.Message) ([]pipeline.Message, error) {
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
	calcBalance := totalAccrual - totalWithdraw

	data.Balance.ResponseBalance = repository.ResponseBalance{Current: calcBalance, Withdrawn: totalWithdraw}

	return []pipeline.Message{data}, nil
}
