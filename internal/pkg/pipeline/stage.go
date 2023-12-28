package pipeline

import (
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
func (chb *StageWorker) Start(done chan struct{}) error {
	for i := 0; i < chb.maxWorkers; i++ {
		//Увеличиваем группу синхронизации на 1
		chb.wg.Add(1)
		// Запускаем в горотине метод по обработке данных, который реализует итерфейс StagePipe{Process(step Message) ([]Message, error)}
		go func() {
			//Уменьшаем группу синхронизации на 1
			defer func() {
				fmt.Println("Close DONE")
				chb.wg.Done()
			}()

			for {
				select {
				// Ожидаем данные поступившие на вход (input)
				case runProcess, ok := <-chb.Input():
					if ok {
						//Запускаем обработку данных передав в  StagePipe{Process()} данные поступившие на вход (input)
						result, err := chb.pipe.Process(runProcess)
						if err != nil {
							log.Println(err)
							close(done)
							return
						}
						for _, r := range result {
							chb.Output() <- r
						}
						//Выполнили задачу и вывалились из горутины
						return
					}
				case <-done:
					fmt.Println("close channel out gorutine")
					return
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
