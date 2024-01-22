package login

import (
	"fmt"
	"sync"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/gzipandunserialize"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/password"
)

// func LoginRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}, sync *sync.Mutex) error {
func LoginRun(dataService *service.Data, sync *sync.Mutex) (chan struct{}, error) {

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
		return nil, err
	}

	data := dataService.GetNewService()

	defaultSet := &data.Default
	//getMetrics := data.GetMetrics

	// Отправялем данные в пайплайн для обработки
	p.Input() <- data
	// Закрываем канал там где отправляем данные
	close(p.Input())

	t1 := time.Now()

	go func() {
		defer func() {
			t2 := time.Now()
			diff := t2.Sub(t1)
			//Выводим время затраченное на выполнение процесса
			fmt.Println("End Http Login:", diff)
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
