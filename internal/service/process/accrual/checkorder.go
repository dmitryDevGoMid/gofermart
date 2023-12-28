package accrual

import (
	"fmt"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type HandlerAccrualCheckOrder struct{}

// Обрабатываем поступивший
func (m HandlerAccrualCheckOrder) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	accrual := &repository.Accrual{}

	accrual.IdOrder = string(data.Default.Body)

	fmt.Println(*accrual)

	/*user, err := data.Default.Repository.SelectUserByEmail(data.Default.Ctx.Request.Context(), &data.User.User)

	//Инициализируем ошибку для ответа клиенту
	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}

		return []pipeline.Message{data}, err
	}

	data.User.User = *user*/

	return []pipeline.Message{data}, nil
}
