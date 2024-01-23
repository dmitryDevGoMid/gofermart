package getlistallorders

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/opentracing/opentracing-go"
)

type HandlerGetListAllOrdersByAccrual struct{}

// Обрабатываем поступившие данные
func (m HandlerGetListAllOrdersByAccrual) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerGetListAllOrders")

	span, ctx := opentracing.StartSpanFromContext(ctx, "Service.Process.HandlerGetListAllOrdersByAccrual")
	defer span.Finish()

	data := result.(*service.Data)

	listOrders, err := data.Default.Repository.SelectAccrualByUser(ctx, &data.User.User)

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
