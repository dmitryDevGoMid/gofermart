package pipeline2

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
		//Увеличиваем группу синхронизации на 1 - в данной реализации pipe&workpool в этом нет необходимости
		chb.wg.Add(1)
		// Запускаем в горотине метод по обработке данных, который реализует итерфейс StagePipe{Process(step Message) ([]Message, error)}
		go func() {
			defer func() {
				fmt.Println("Close gorutines...")
				chb.wg.Done()
				//Закрываем входящий канал для следующего pipe - данных больше не будет! (цепная реакция, которая закроет все горутины)
				//close(chb.Output())
			}()

			for runProcess := range chb.Input() {
				result, err := chb.pipe.Process(runProcess)
				if err != nil {
					log.Println(err)
					continue
				}
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
