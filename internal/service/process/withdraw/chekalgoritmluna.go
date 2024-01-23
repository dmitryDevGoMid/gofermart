package withdraw

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/luna"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/opentracing/opentracing-go"
)

type WithDrawCheckAlgoritmLuna struct{}

// Обрабатываем поступивший
func (m WithDrawCheckAlgoritmLuna) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	fmt.Println("Execute HandlerAccrual")

	span, _ := opentracing.StartSpanFromContext(ctx, "Service.Process.WithDrawCheckAlgoritmLuna")
	defer span.Finish()

	data := result.(*service.Data)

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
