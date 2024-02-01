package loyalty

import (
	"context"

	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

// Обрабатываем поступивший
func PrepeareDataByAccrual(ctx context.Context, data *service.Data) error {

	data.Loyalty.Accruals = &[]repository.Accrual{}

	dataLinkAccrual, err := data.Default.Repository.SelectAccrualForSendLoyalty(ctx, data.Loyalty.Accruals)

	if err != nil {
		return err
	}

	//Если нет ошибок значит есть данные сохраняем ссылку в структуре
	data.Loyalty.Accruals = dataLinkAccrual

	return nil
}
