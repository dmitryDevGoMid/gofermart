package loyalty

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type ResponseLoyalty struct{}

// Обрабатываем поступивший
func (m ResponseLoyalty) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	err := json.Unmarshal(data.Loyalty.Response, &data.Loyalty.ResponseLoyaltyService)
	if err != nil {
		return []pipeline.Message{data}, errors.New("ResponseLoyalty " + err.Error())
	}

	return []pipeline.Message{data}, nil
}
