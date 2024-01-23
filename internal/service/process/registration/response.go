package registration

import (
	"context"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

type ResponseRegistration struct{}

// Обрабатываем поступивший
func (m ResponseRegistration) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {

	span, _ := opentracing.StartSpanFromContext(ctx, "Service.Process.ResponseRegistration")
	defer span.Finish()

	data := result.(*service.Data)

	data.Default.Response = func() {
		data.Default.Ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": string("Success"), // cast it to string before showing
		})
	}

	return []pipeline.Message{data}, nil
}
