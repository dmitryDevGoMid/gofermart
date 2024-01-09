package loyalty

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

//Запускаем pipeline для процесса регистрации клиента в сервисе

func LoyaltyRun(ctx context.Context, cfg *config.Config, repository repository.Repository, ticker *time.Ticker) error {

	t1 := time.Now()

	ticker.Stop()

	defer func() {
		fmt.Println("Ticker-Time Set", cfg.TickerTime.TickTack)
		//Устанавливаем новое значение (возможен ответ 429 для )
		ticker.Reset(time.Duration(cfg.TickerTime.TickTack) * time.Second)
		//Сбрасываем в значение по умолчанию 3 - sec
		cfg.TickerTime.TickTack = 3
	}()

	data := &service.Data{}

	//Устанавливаем конфигурационные данные
	data.Default.Cfg = cfg
	//Устанавливаем репозитарий для данных из базы
	data.Default.Repository = repository

	//Подготавливаем заказы для запроса к системе лояльности
	err := PrepeareDataByAccrual(ctx, data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	p := pipeline.NewConcurrentPipeline()

	//Отправляем данные в потоке в лояльность
	p.AddPipe(RequestLoyalty{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Обрабатываем ответ
	p.AddPipe(ResponseLoyalty{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	//Изменяем данные
	p.AddPipe(CahngeDataByResponseLoyalty{}, &pipeline.PipelineOpts{
		MaxWorkers: 1,
	})

	if err := p.Start(); err != nil {
		log.Println(err)
	}

	//Здесь формируем отдельную область памяти для каждого начисления и отправляем в канал
	go func() {
		defer close(p.Input())

		for _, val := range *data.Loyalty.Accruals {

			data.Loyalty.Accrual = val

			//Разименовываем указатель и отправляем в канал с указателями и репозитарий и индивидуальными данными по начислениям
			dataForSend := *data

			//log.Println("Пишем в канал данные:", dataForSend)
			select {
			case <-ctx.Done():
				return
			default:
				p.Input() <- &dataForSend
			}
		}
	}()

	go func() {
		count := 0
		defer func() {
			//datas.err()
			fmt.Println("End of Loyalty!End of Loyalty!End of Loyalty!End of Loyalty!End of Loyalty!End of Loyalty!End of Loyalty!")
			t2 := time.Now()
			diff := t2.Sub(t1)
			fmt.Println(diff)
			//close(p.Done())
		}()
		for defaultSet := range p.Output() {
			data := defaultSet.(*service.Data)
			data.Default.Response()
			fmt.Println(count)
			count++
		}
		/*for {
			select {
			case defaultSet, ok := <-p.Output():
				if ok {
					data := defaultSet.(service.Data)
					data.Default.Response()
					//fmt.Println("Получили данные на выходе:", data)
					fmt.Println(count)
					count++
				}
			case <-ctx.Done():
				//close(p.Input())
				return
			default:
			}
		}*/
	}()

	p.Stop()

	//time.Sleep(time.Duration(5) * time.Second)
	//p.Stop()

	return nil
}
