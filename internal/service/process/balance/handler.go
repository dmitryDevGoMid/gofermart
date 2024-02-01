package balance

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerBalance struct{}

// Обрабатываем поступившие данные
func (m HandlerBalance) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.HandlerBalance")
	if span != nil {
		defer span.Finish()
	}

	//Сумма списаний
	totalWithdraw, err := data.Default.Repository.SelectWithdrawByUserSum(ctx, &data.User.User)
	fmt.Println(totalWithdraw)
	if err != nil {

		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusInternalServerError)
		}

		return []pipeline.Message{data}, err
	}

	//Сумма начислений
	totalAccrual, err := data.Default.Repository.SelectAccrualByUserSum(ctx, &data.User.User)
	fmt.Println(totalAccrual)
	if err != nil {

		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusInternalServerError)
		}

		return []pipeline.Message{data}, err
	}

	//Разница
	calcBalance := totalAccrual - totalWithdraw

	data.Balance.ResponseBalance = repository.ResponseBalance{Current: math.Round(float64(calcBalance)*100) / 100, Withdrawn: math.Round(float64(totalWithdraw)*100) / 100}

	return []pipeline.Message{data}, nil
}
