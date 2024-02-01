package login

import (
	"context"
	"errors"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerLogin struct{}

// Обрабатываем поступивший
func (m HandlerLogin) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.HandlerLogin")
	if span != nil {
		defer span.Finish()
	}

	user, err := data.Default.Repository.SelectUserByEmail(ctx, &data.User.User)

	//Инициализируем ошибку для ответа клиенту
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusInternalServerError)
		}

		return []pipeline.Message{data}, err
	}

	if user == nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusUnauthorized)
		}

		return []pipeline.Message{data}, errors.New("user is not authorized")
	}

	data.User.User = *user

	return []pipeline.Message{data}, nil
}
