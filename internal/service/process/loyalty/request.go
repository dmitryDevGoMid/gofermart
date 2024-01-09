package loyalty

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/go-resty/resty/v2"
)

type RequestLoyalty struct{}

// Обрабатываем поступивший
func (m RequestLoyalty) Process(result pipeline.Message) ([]pipeline.Message, error) {
	data := result.(*service.Data)

	client := resty.New()

	//fmt.Println(data.Accrual.Accrual.IDorder)
	urlMetrics := fmt.Sprintf("%s/api/orders/%s", data.Default.Cfg.AccrualSystem.AccrualSystem, data.Loyalty.Accrual.IDorder)

	//fmt.Println(urlMetrics)
	//fmt.Println(urlMetrics)
	response, err := client.R().Get(urlMetrics)
	if err != nil {
		fmt.Println(err)
	}

	//Выполняем действия в зависимости от стутатуса
	switch response.StatusCode() {
	case 500, 204:
		errString := fmt.Sprintf(`loyalty response %d`, response.StatusCode())
		return []pipeline.Message{data}, errors.New(errString)
	case 429:
		retryTimeSecond := response.Header().Get("Retry-After")
		timeTicker, err := strconv.Atoi(retryTimeSecond)
		//Если есть ошибка просто выставляем по умочанию тикер 60-seconds
		if err != nil {
			fmt.Println(err)
			data.Default.Cfg.TickerTime.TickTack = 60
		} else {
			data.Default.Cfg.TickerTime.TickTack = timeTicker
		}
		errString := fmt.Sprintf(`loyalty response %d`, response.StatusCode())
		return []pipeline.Message{data}, errors.New(errString)
	}

	fmt.Println("LOYLTY RESPONSE BODY=====>", response)

	response.StatusCode()

	data.Loyalty.Response = response.Body()

	data.Default.Response = func() {
		//fmt.Println("LOYLTY RESPONSE BODY=====>", response)
	}

	return []pipeline.Message{data}, nil
}
