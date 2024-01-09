package accrual

import (
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderCRUDAccrual struct{}

// Обрабатываем поступивший
func (m OrderCRUDAccrual) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	data.Accrual.Accrual.IDUser = data.User.User.ID

	err := data.Default.Repository.InsertAccrual(data.Default.Ctx.Request.Context(), &data.Accrual.Accrual)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": string("Trable InsertAccrual"), // cast it to string before showing
			})
		}

		return []pipeline.Message{data}, err
	}

	data.Default.Response = func() {
		data.Default.Ctx.Status(202)
	}

	return []pipeline.Message{data}, nil
}
