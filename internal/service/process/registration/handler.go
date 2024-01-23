package registration

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/opentracing/opentracing-go"
)

type HandlerRegistration struct{}

// Обрабатываем поступивший
func (m HandlerRegistration) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Processing HandlerRegistration")

	span, ctx := opentracing.StartSpanFromContext(ctx, "Service.Process.HandlerRegistration")
	defer span.Finish()

	data := result.(*service.Data)

	user, err := data.Default.Repository.SelectUserByEmail(ctx, &data.User.User)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
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
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err

	}

	data.User.User = *user

	data.Default.Response = func() {
		data.Default.Ctx.JSON(http.StatusOK, data.User.User)
	}

	return []pipeline.Message{data}, nil
}
