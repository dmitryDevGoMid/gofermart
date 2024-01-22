package registration

import (
	"fmt"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/authentication"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/gzipandunserialize"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/password"
)

type User struct {
	User         repository.User
	UserRequest  repository.User
	HashPassword string
}

//Запускаем pipeline для процесса регистрации клиента в сервисе

// func RegistrationRun(ctx context.Context, c *gin.Context, cfg *config.Config, rep repository.Repository, finished chan struct{}, sync *sync.Mutex) error {
func RegistrationRun(dataService *service.Data) (chan struct{}, error) {

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

	p.AddPipe(ResponseRegistration{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	if err := p.Start(); err != nil {
		return nil, err
	}

	//Получаем структуру данных для работы pipeline
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
			fmt.Println("End Http Registered:", diff)

			//Отсавляю один метод на все, через который отдаем как успех так фэйл ()
			defaultSet.Response()

			// Закрываем канал с пустой структурой сигнал, о завершении выполнения pipeline
			close(defaultSet.Finished)
		}()

		//Ожидаем данные из канала output
		<-p.Output()
	}()

	//
	p.Stop()

	return defaultSet.Finished, nil
}
