package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"
)

//##################Stage(Workers для каждого Pipe отдельно)#################

type StageWorker struct {
	wg         *sync.WaitGroup
	input      chan Message
	output     chan Message
	maxWorkers int
	pipe       StagePipe
}

// Воркеры для каждого pipe свои кол-во определяется в PipelineOpts{maxWorkers}
func (chb *StageWorker) Start(ctx context.Context, done chan struct{}) error {
	for i := 0; i < chb.maxWorkers; i++ {
		//Увеличиваем группу синхронизации на 1 - в данной реализации pipe&workpool
		chb.wg.Add(1)
		// Запускаем в горотине метод по обработке данных, который реализует итерфейс StagePipe{Process(ctx context.Context,step Message) ([]Message, error)}
		go func() {
			defer func() {
				fmt.Println("Close gorutines...")
				chb.wg.Done()
			}()

			for runProcess := range chb.Input() {
				result, err := chb.pipe.Process(ctx, runProcess)
				if err != nil {
					log.Println(err)
					continue
				}
				//Передаем данные в output канал текущего pipe, который является input для следующего
				for _, r := range result {
					chb.Output() <- r
				}
			}
		}()
	}

	return nil

}

// Запускаем процесс ожидания завершения работников
func (chb *StageWorker) WaitStop() error {
	chb.wg.Wait()
	fmt.Print("Close ALLL")
	return nil
}

// Возвращаем канал входных данных
func (chb *StageWorker) Input() chan Message {
	return chb.input
}

// Возвращаем канал выходных данных
func (chb *StageWorker) Output() chan Message {
	return chb.output
}

// Возвращает проинициализированную гуппу работников готовых к обработке данных (запуск через метод Start->StageWork)
func NewWorkerGroup(maxWorkers int, pipe StagePipe, input chan Message, output chan Message) StageWorker {
	return StageWorker{
		wg:         &sync.WaitGroup{},
		input:      input,
		output:     output,
		maxWorkers: maxWorkers,
		pipe:       pipe,
	}
}
