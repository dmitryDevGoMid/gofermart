package accrual

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/gzipandunserialize"
	"github.com/gin-gonic/gin"
)

func AccrualRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}, sync *sync.Mutex) error {

	sync.Lock()

	defer func() {
		sync.Unlock()
	}()

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

	if err := p.Start(); err != nil {
		log.Println(err)
	}

	data := &service.Data{}

	//Устанавливаем контекст запроса gin
	data.Default.Ctx = c
	//Устанавливаем конфигурационные данные
	data.Default.Cfg = cfg
	//Устанавливаем канал завершения процесса
	data.Default.Finished = finished
	//Устанавливаем репозитарий для данных из базы
	data.Default.Repository = rep

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
	return nil
}
