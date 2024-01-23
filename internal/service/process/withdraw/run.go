package withdraw

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/gzipandunserialize"
)

func WithdrawRun(ctx context.Context, dataService *service.Data, sync *sync.Mutex) (chan struct{}, error) {
	sync.Lock()

	defer sync.Unlock()

	p := pipeline.NewConcurrentPipeline()

	//Проверяем наличие токена
	p.AddPipe(authentication.CheckJWTToken{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Если необходимо ungzip body
	p.AddPipe(gzipandunserialize.Gzip{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(gzipandunserialize.UnserializeWithdraw{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем тело запроса в стек с данными body
	p.AddPipe(WithDrawCheckAlgoritmLuna{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем тело запроса в стек с данными body
	p.AddPipe(HandlerWithdraw{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Сохраняем тело запроса в стек с данными body
	p.AddPipe(OrderCRUDWithdraw{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	if err := p.Start(ctx); err != nil {
		return nil, err
	}

	data := dataService.GetNewService()

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
