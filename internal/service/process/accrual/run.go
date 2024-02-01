package accrual

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/check"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/gzipandunserialize"
)

func AccrualRun(ctx context.Context, dataService *service.Data, sync *sync.Mutex) (chan struct{}, error) {

	sync.Lock()

	defer func() {
		sync.Unlock()
	}()

	data := dataService.GetNewService()

	span, _ := data.Default.Tracing.Tracing(ctx, "Service.Process.AccrualRun")
	if span != nil {
		defer span.Finish()
	}

	p := pipeline.NewConcurrentPipeline()

	//Проверяем ContentType запроса
	p.AddPipe(check.CheckContentTypeOrders{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Проверяем наличие токена
	p.AddPipe(authentication.CheckJWTToken{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Если необходимо ungzip body
	p.AddPipe(gzipandunserialize.Gzip{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Проверяем номер заказа по алгоритму luna
	p.AddPipe(AccrualCheckAlgoritmLuna{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Обращаемся к базе и проверяем наличие заказа
	p.AddPipe(HandlerAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем запрос в базе
	p.AddPipe(OrderCRUDAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Формируем успешный ответ клиенту
	p.AddPipe(ResponseAccrual{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	if err := p.Start(ctx); err != nil {
		return nil, err
	}

	defaultSet := &data.Default

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
