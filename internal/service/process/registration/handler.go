package registration

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerRegistration struct{}

// Обрабатываем поступивший
func (m HandlerRegistration) Process(result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Processing HandlerRegistration")

	data := result.(*service.Data)

	user, err := data.Default.Repository.SelectUserByEmail(data.Default.Ctx.Request.Context(), &data.User.User)

	fmt.Println(data.User.User)

	if user != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusConflict)
		}
		return []pipeline.Message{data}, errors.New("This login already exists")
	}

	user, err = data.Default.Repository.InsertUser(data.Default.Ctx.Request.Context(), &data.User.User)

	//Инициализируем ошибку для ответа клиенту
	if err != nil {
		data.Default.ResponseError = func() {
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
