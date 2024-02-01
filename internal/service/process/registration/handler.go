package registration

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerRegistration struct{}

// Обрабатываем поступивший
func (m HandlerRegistration) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Processing HandlerRegistration")

	data := result.(*service.Data)

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.HandlerRegistration")
	if span != nil {
		defer span.Finish()
	}

	user, err := data.Default.Repository.SelectUserByEmail(ctx, &data.User.User)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusInternalServerError)
		}
		return []pipeline.Message{data}, err

	}

	if user != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusConflict)
		}
		return []pipeline.Message{data}, errors.New("this login already exists")
	}

	user, err = data.Default.Repository.InsertUser(ctx, &data.User.User)

	//Инициализируем ошибку для ответа клиенту
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusInternalServerError)
		}
		return []pipeline.Message{data}, err

	}

	data.User.User = *user

	data.Default.Response = func() {
		data.Default.Ctx.JSON(http.StatusOK, data.User.User)
	}

	return []pipeline.Message{data}, nil
}
