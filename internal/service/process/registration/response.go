package registration

import (
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/gin-gonic/gin"
)

type ResponseRegistration struct{}

// Обрабатываем поступивший
func (m ResponseRegistration) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	data.Default.Response = func() {
		data.Default.Ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": string("Success"), // cast it to string before showing
		})
	}

	return []pipeline.Message{data}, nil
}