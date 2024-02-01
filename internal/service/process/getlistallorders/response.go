package getlistallorders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type ResponseGetListAllOrdersByAccrual struct{}

// Обрабатываем поступивший
func (m ResponseGetListAllOrdersByAccrual) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	span, _ := data.Default.Tracing.Tracing(ctx, "Service.Process.ResponseGetListAllOrdersByAccrual")
	if span != nil {
		defer span.Finish()
	}

	dataResponse, err := json.Marshal(data.Accrual.AccrualList)

	fmt.Println(dataResponse)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, nil
	}

	data.Default.Response = func() {
		data.Default.Ctx.Data(http.StatusOK, "application/json", dataResponse)
	}

	return []pipeline.Message{data}, nil
}
