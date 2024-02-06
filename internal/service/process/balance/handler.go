package balance

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
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

	data.User.PbUser = new(pb.User)

	data.User.PbUser.Id = int32(data.User.User.ID)
	data.User.PbUser.Name = data.User.User.Login
	data.User.PbUser.Countorders = 5

	//Выполняем gRPC запрос для получения дополнительного бонуса
	resp, err := data.Default.PbCleint.GetBonusPlus(ctx, data.User.PbUser)
	if err != nil {
		fmt.Println("error getBonusPlus: ", err)
	}

	fmt.Println("BonusPlus:", resp.Bonus)
	data.User.BonusPlus = resp.Bonus

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

	//Разница + дополнительный бонус полученный по gRPC по кол-ву заказов
	calcBalance := (totalAccrual - totalWithdraw) + data.User.BonusPlus

	data.Balance.ResponseBalance = repository.ResponseBalance{Current: math.Round(float64(calcBalance)*100) / 100, Withdrawn: math.Round(float64(totalWithdraw)*100) / 100, Bonusplus: math.Round(float64(data.User.BonusPlus)*100) / 100}
	return []pipeline.Message{data}, nil
}
