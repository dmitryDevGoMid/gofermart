package registration

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
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/password"
	"github.com/gin-gonic/gin"
)

type User struct {
	User         repository.User
	UserRequest  repository.User
	HashPassword string
}

//Запускаем pipeline для процесса регистрации клиента в сервисе

func RegistrationRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}, sync *sync.Mutex) error {

	sync.Lock()

	defer sync.Unlock()

	p := pipeline.NewConcurrentPipeline()

	p.AddPipe(gzipandunserialize.Gzip{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(gzipandunserialize.UnserializeUser{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(password.SetHashPassword{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(HandlerRegistration{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(authentication.SetJWTToken{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	p.AddPipe(authentication.CheckJWTToken{}, &pipeline.PipelineOpts{
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
