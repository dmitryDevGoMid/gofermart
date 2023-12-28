package withdraw

import (
	"errors"
	"fmt"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/luna"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type WithDrawCheckAlgoritmLuna struct{}

// Обрабатываем поступивший
func (m WithDrawCheckAlgoritmLuna) Process(result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	err := luna.Validate(data.Withdraw.Withdraw.IDorder)

	//Проверяем номер на валидность
	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(422)
		}
		return []pipeline.Message{data}, errors.New("Invalid check order number withdraw")
	}

	return []pipeline.Message{data}, nil
}
