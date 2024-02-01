package withdraw

import (
	"context"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderCRUDWithdraw struct{}

// Обрабатываем поступивший
func (m OrderCRUDWithdraw) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.OrderCRUDWithdraw")
	if span != nil {
		defer span.Finish()
	}

	data.Withdraw.Withdraw.IDUser = data.User.User.ID

	err := data.Default.Repository.InsertWithdraw(ctx, &data.Withdraw.Withdraw)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": string("Trable StatusInternalServerError"), // cast it to string before showing
			})
		}

		return []pipeline.Message{data}, err
	}

	data.Default.Response = func() {
		data.Default.Ctx.Status(200)
	}

	return []pipeline.Message{data}, nil
}
