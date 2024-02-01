package password

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/security"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type SetHashPassword struct{}

// func (chain *CheckGzip) run(r *Request) error {
func (m SetHashPassword) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	var err error

	data.User.HashPassword, err = security.EncryptPassword(data.User.UserRequest.Password)

	//Инициализируем ошибку для ответа клиенту
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err

	}

	if !data.Default.Cfg.Server.TestingEnabled {
		data.User.User.Password = data.User.HashPassword
	}

	fmt.Println(data.User.User)

	return []pipeline.Message{data}, nil
}
