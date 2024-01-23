package accrual

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/gzipandunserialize"
	"github.com/opentracing/opentracing-go"
)

func AccrualRun(ctx context.Context, dataService *service.Data, sync *sync.Mutex) (chan struct{}, error) {

	sync.Lock()

	defer func() {
		sync.Unlock()
	}()

	span, ctx := opentracing.StartSpanFromContext(ctx, "Service.Process.AccrualRun")
	defer span.Finish()

	p := pipeline.NewConcurrentPipeline()

	//Проверяем наличие токена
	p.AddPipe(authentication.CheckJWTToken{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Если необходимо ungzip body
	p.AddPipe(gzipandunserialize.Gzip{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем тело запроса в стек с данными body
	p.AddPipe(AccrualCheckAlgoritmLuna{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем тело запроса в стек с данными body
	p.AddPipe(HandlerAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем тело запроса в стек с данными body
	p.AddPipe(OrderCRUDAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(ResponseAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	if err := p.Start(ctx); err != nil {
		return nil, err
	}

	data := dataService.GetNewService()

	defaultSet := &data.Default
	//getMetrics := data.GetMetrics

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
			fmt.Println("End Http Accrual:", diff)
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
