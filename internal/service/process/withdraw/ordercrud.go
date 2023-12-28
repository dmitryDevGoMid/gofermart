package withdraw

import (
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderCRUDWithdraw struct{}

// Обрабатываем поступивший
func (m OrderCRUDWithdraw) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	data.Withdraw.Withdraw.IdUser = data.User.User.Id

	err := data.Default.Repository.InsertWithdraw(data.Default.Ctx.Request.Context(), &data.Withdraw.Withdraw)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": string("Trable InsertWithdraw"), // cast it to string before showing
			})
		}

		return []pipeline.Message{data}, err
	}

	data.Default.Response = func() {
		data.Default.Ctx.Status(202)
	}

	return []pipeline.Message{data}, nil
}
