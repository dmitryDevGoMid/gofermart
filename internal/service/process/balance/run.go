package balance

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline2"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication2"
	"github.com/gin-gonic/gin"
)

func BalanceRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}, sync *sync.Mutex) error {

	sync.Lock()

	defer sync.Unlock()

	//p := pipeline.NewConcurrentPipeline()
	p := pipeline2.NewConcurrentPipeline()
	//Проверяем наличие токена
	p.AddPipe(authentication2.CheckJWTToken{}, &pipeline2.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(HandlerBalance{}, &pipeline2.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(ResponseBalance{}, &pipeline2.PipelineOpts{
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
	data.Default.Ctx.Writer.Header().Set("Content-Type", "application/json")

	//Устанавливаем канал завершения процесса
	data.Default.Finished = finished

	//Устанавливаем репозитарий для данных из базы
	data.Default.Repository = rep

	defaultSet := &data.Default

	//Отправялем данные в пайплайн для обработки
	p.Input() <- data
	//Закрываем канал там где отправляем данные
	close(p.Input())

	t1 := time.Now()

	go func() {
		defer func() {
			//datas.err()
			fmt.Println("End Http GetBalance:")
			t2 := time.Now()
			diff := t2.Sub(t1)
			fmt.Println(diff)
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
