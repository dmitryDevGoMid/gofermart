package login

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
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/password"
	"github.com/gin-gonic/gin"
)

func LoginRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}, sync *sync.Mutex) error {

	sync.Lock()

	defer sync.Unlock()

	p := pipeline.NewConcurrentPipeline()

	p.AddPipe(gzipandunserialize.Gzip{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(gzipandunserialize.UnserializeLogin{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(HandlerLogin{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(password.Authetication{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(authentication.SetJWTToken{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(ResponseLogin{}, &pipeline.PipelineOpts{
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

	// Отправялем данные в пайплайн для обработки
	p.Input() <- data
	// Закрываем канал там где отправляем данные
	close(p.Input())

	t1 := time.Now()

	go func() {
		defer func() {
			//datas.err()
			fmt.Println("End Http Login:")
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
