package balance

import (
	"fmt"
	"math"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline2"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerBalance struct{}

// Обрабатываем поступившие данные
func (m HandlerBalance) Process(result pipeline2.Message) ([]pipeline2.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	//Сумма списаний
	totalWithdraw, err := data.Default.Repository.SelectWithdrawByUserSum(data.Default.Ctx.Request.Context(), &data.User.User)
	fmt.Println(totalWithdraw)
	if err != nil {
		/*data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}*/

		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline2.Message{data}, err
	}

	//Сумма начислений
	totalAccrual, err := data.Default.Repository.SelectAccrualByUserSum(data.Default.Ctx.Request.Context(), &data.User.User)
	fmt.Println(totalAccrual)
	if err != nil {
		/*data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}*/

		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline2.Message{data}, err
	}

	//Разница
	calcBalance := totalAccrual - totalWithdraw

	data.Balance.ResponseBalance = repository.ResponseBalance{Current: math.Round(float64(calcBalance)*100) / 100, Withdrawn: math.Round(float64(totalWithdraw)*100) / 100}

	return []pipeline2.Message{data}, nil
}
