package withdraw

import (
	"context"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

type OrderCRUDWithdraw struct{}

// Обрабатываем поступивший
func (m OrderCRUDWithdraw) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Service.Process.HandlerWithdraw")
	defer span.Finish()

	data := result.(*service.Data)

	data.Withdraw.Withdraw.IDUser = data.User.User.ID

	err := data.Default.Repository.InsertWithdraw(ctx, &data.Withdraw.Withdraw)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": string("Trable InsertWithdraw"), // cast it to string before showing
			})
		}

		return []pipeline.Message{data}, err
	}

	data.Default.Response = func() {
		data.Default.Ctx.Status(200)
	}

	return []pipeline.Message{data}, nil
}
