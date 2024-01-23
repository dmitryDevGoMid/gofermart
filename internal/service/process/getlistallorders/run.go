package getlistallorders

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
)

//Запускаем pipeline для процесса регистрации клиента в сервисе

// func GetAllListOrtdersRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}) error {

func GetAllListOrtdersRun(ctx context.Context, dataService *service.Data) (chan struct{}, error) {
	p := pipeline.NewConcurrentPipeline()

	//Проверяем наличие токена
	p.AddPipe(authentication.CheckJWTToken{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Проверяем наличие токена
	p.AddPipe(HandlerGetListAllOrdersByAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Проверяем наличие токена
	p.AddPipe(ResponseGetListAllOrdersByAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	if err := p.Start(ctx); err != nil {
		return nil, err
	}

	data := dataService.GetNewService()

	defaultSet := &data.Default
	data.Default.Ctx.Writer.Header().Set("Content-Type", "application/json")

	//Отправялем данные в пайплайн для обработки
	p.Input() <- data
	//Закрываем канал там где отправляем данные
	close(p.Input())

	t1 := time.Now()

	go func() {
		defer func() {
			t2 := time.Now()
			diff := t2.Sub(t1)
			//Выводим время затраченное на выполнение процесса
			fmt.Println("End Http GeyListAllOrders:", diff)
			//Отсавляю один метод на все через, который отдаем как успех так фэйл
			defaultSet.Response()
			close(defaultSet.Finished)
		}()

		//Ожидаем данные из канала output
		<-p.Output()
	}()

	p.Stop()

	return defaultSet.Finished, nil
}
