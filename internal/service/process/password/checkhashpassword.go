package password

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/security"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type Authetication struct{}

// func (chain *CheckGzip) run(r *Request) error {
func (m Authetication) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	err := security.VerifyPassword(data.User.User.Password, data.User.UserRequest.Password)
	fmt.Println(err)

	//Инициализируем ошибку для ответа клиенту
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err

	}

	return []pipeline.Message{data}, nil
}
