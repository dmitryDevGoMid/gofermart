package accrual

import (
	"errors"
	"fmt"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/luna"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type AccrualCheckAlgoritmLuna struct{}

// Обрабатываем поступивший
func (m AccrualCheckAlgoritmLuna) Process(result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	accrual := &repository.Accrual{}

	accrual.IDorder = string(data.Default.Body)

	err := luna.Validate(accrual.IDorder)

	//Проверяем номер на валидность
	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(422)
		}
		return []pipeline.Message{data}, errors.New("invalid check number")
	}

	data.Accrual.Accrual = *accrual

	return []pipeline.Message{data}, nil
}
