package loyalty

import (
	"context"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline2"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type CahngeDataByResponseLoyalty struct{}

// Обрабатываем поступивший
func (m CahngeDataByResponseLoyalty) Process(result pipeline2.Message) ([]pipeline2.Message, error) {
	data := result.(*service.Data)

	catalogData := data.Default.Repository.GetCatalogData(context.TODO())

	data.Loyalty.Accrual.IdStatus = catalogData.TypeStatus[data.Loyalty.ResponseLoyaltyService.Status]
	data.Loyalty.Accrual.Accrual = data.Loyalty.ResponseLoyaltyService.Accrual

	err := data.Default.Repository.UpdateAccrualById(context.TODO(), &data.Loyalty.Accrual)

	if err != nil {
		return []pipeline2.Message{data}, err
	}

	return []pipeline2.Message{data}, nil
}
