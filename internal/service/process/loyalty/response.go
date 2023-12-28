package loyalty

import (
	"encoding/json"
	"errors"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline2"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type ResponseLoyalty struct{}

// Обрабатываем поступивший
func (m ResponseLoyalty) Process(result pipeline2.Message) ([]pipeline2.Message, error) {
	data := result.(*service.Data)

	//data.Loyalty.Response

	err := json.Unmarshal(data.Loyalty.Response, &data.Loyalty.ResponseLoyaltyService)
	if err != nil {
		return []pipeline2.Message{data}, errors.New("ResponseLoyalty " + err.Error())
	}

	return []pipeline2.Message{data}, nil
}
