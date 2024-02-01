package withdraw

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

func WithdrawRun(ctx context.Context, dataService *service.Data, sync *sync.Mutex) (chan struct{}, error) {
	sync.Lock()

	defer sync.Unlock()

	data := dataService.GetNewService()

	span, ctx := data.Default.Tracing.Tracing(ctx, "Service.Process.WithdrawRun")
	if span != nil {
		defer span.Finish()
	}

	p := pipeline.NewConcurrentPipeline()

	//Проверяем Content-type
	p.AddPipe(check.CheckContentTypeWithdraw{}, &pipeline.PipelineOpts{
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

	//Выполняем унсерализацию данных
	p.AddPipe(gzipandunserialize.UnserializeWithdraw{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Проверяем номер заказа
	p.AddPipe(WithDrawCheckAlgoritmLuna{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Выполняем проверку на возможность списания по остатоку
	p.AddPipe(HandlerWithdraw{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем заказа на списание
	p.AddPipe(OrderCRUDWithdraw{}, &pipeline.PipelineOpts{
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
			fmt.Println("End Http Withdraw:", diff)
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
