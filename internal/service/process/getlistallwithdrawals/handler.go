package getlistallwithdrawals

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/opentracing/opentracing-go"
)

type HandlerGetListAllOrdersByWithDraw struct{}

// Обрабатываем поступившие данные
func (m HandlerGetListAllOrdersByWithDraw) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerGetListAllOrdersByWithDraw")

	span, ctx := opentracing.StartSpanFromContext(ctx, "Service.Process.HandlerGetListAllOrdersByWithDraw")
	defer span.Finish()

	data := result.(*service.Data)

	listWithdrawals, err := data.Default.Repository.SelectWithdrawByUsers(ctx, &data.User.User)

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
