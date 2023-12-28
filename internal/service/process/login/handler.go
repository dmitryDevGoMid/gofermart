package login

import (
	"errors"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerLogin struct{}

// Обрабатываем поступивший
func (m HandlerLogin) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	user, err := data.Default.Repository.SelectUserByEmail(data.Default.Ctx.Request.Context(), &data.User.User)

	//Инициализируем ошибку для ответа клиенту
	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	if user == nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusUnauthorized)
		}

		return []pipeline.Message{data}, errors.New("user is not authorized")
	}

	data.User.User = *user

	return []pipeline.Message{data}, nil
}
