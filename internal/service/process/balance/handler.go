package balance

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/opentracing/opentracing-go"
)

type HandlerBalance struct{}

// Обрабатываем поступившие данные
func (m HandlerBalance) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	span, ctx := opentracing.StartSpanFromContext(ctx, "Service.Process.HandlerBalance")
	defer span.Finish()

	data := result.(*service.Data)

	//Сумма списаний
	totalWithdraw, err := data.Default.Repository.SelectWithdrawByUserSum(ctx, &data.User.User)
	fmt.Println(totalWithdraw)
	if err != nil {

		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	//Сумма начислений
	totalAccrual, err := data.Default.Repository.SelectAccrualByUserSum(ctx, &data.User.User)
	fmt.Println(totalAccrual)
	if err != nil {

		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	//Разница
	calcBalance := totalAccrual - totalWithdraw

	data.Balance.ResponseBalance = repository.ResponseBalance{Current: math.Round(float64(calcBalance)*100) / 100, Withdrawn: math.Round(float64(totalWithdraw)*100) / 100}

	return []pipeline.Message{data}, nil
}
