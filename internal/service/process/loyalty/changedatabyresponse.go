package loyalty

import (
	"context"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type CahngeDataByResponseLoyalty struct{}

// Обрабатываем поступивший
func (m CahngeDataByResponseLoyalty) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	catalogData := data.Default.Repository.GetCatalogData(context.TODO())

	data.Loyalty.Accrual.IDStatus = catalogData.TypeStatus[data.Loyalty.ResponseLoyaltyService.Status]
	data.Loyalty.Accrual.Accrual = data.Loyalty.ResponseLoyaltyService.Accrual

	err := data.Default.Repository.UpdateAccrualByID(context.TODO(), &data.Loyalty.Accrual)

	if err != nil {
		return []pipeline.Message{data}, err
	}

	return []pipeline.Message{data}, nil
}
