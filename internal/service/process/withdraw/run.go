package withdraw

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/gzipandunserialize"
	"github.com/gin-gonic/gin"
)

func WithdrawRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}, sync *sync.Mutex) error {

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

	p.Input() <- data

	go func() {
		defer func() {
			//datas.err()
			fmt.Println("End of!")
			close(defaultSet.Finished)
		}()
		print := "Hello Worl!"

		for {
			select {
			case _, ok := <-p.Output():
				if ok {
					//data := data.(*service.Data)
					//fmt.Println("=====>!", defaultSet.Metrics)
					//defaultSet.Ctx.Status(http.StatusOK)
					defaultSet.Response()
					return
				}
			case <-p.Done():
				fmt.Println("Close====>Output")
				//defaultSet.Err()
				defaultSet.ResponseError()
				fmt.Println(print)
				return
			}
		}
	}()

	p.Stop()

	return nil
}
