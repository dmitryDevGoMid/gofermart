package getlistallorders

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
)

func GetAllListOrtdersRun(ctx context.Context, dataService *service.Data) (chan struct{}, error) {

	data := dataService.GetNewService()

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.ResponseGetListAllOrdersByAccrual")
	if span != nil {
		defer span.Finish()
	}

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

	//Запускаем пайплайн в работу
	if err := p.Start(ctx); err != nil {
		return nil, err
	}

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
