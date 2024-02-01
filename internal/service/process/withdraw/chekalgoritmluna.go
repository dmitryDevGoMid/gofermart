package withdraw

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/luna"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type WithDrawCheckAlgoritmLuna struct{}

// Обрабатываем поступивший
func (m WithDrawCheckAlgoritmLuna) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	data := result.(*service.Data)

	span, _ := data.Default.Tracing.Tracing(ctx, "Service.Process.WithDrawCheckAlgoritmLuna")
	if span != nil {
		defer span.Finish()
	}

	err := luna.Validate(data.Withdraw.Withdraw.IDorder)

	//Проверяем номер на валидность
	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(422)
		}
		return []pipeline.Message{data}, errors.New("invalid check order number withdraw")
	}

	return []pipeline.Message{data}, nil
}
