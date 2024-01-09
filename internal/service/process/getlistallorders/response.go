package getlistallorders

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type ResponseGetListAllOrdersByAccrual struct{}

// Обрабатываем поступивший
func (m ResponseGetListAllOrdersByAccrual) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

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
