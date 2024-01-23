package pipeline

import (
	"context"
	"errors"
)

// Пустой интерфейс в который кладем данные для движения по pipes
type Message interface {
}

// Итерфейс для реализации pipe (каркас)
type StagePipe interface {
	Process(ctx context.Context, step Message) ([]Message, error)
}

// Интерфейс для организации запуска и передачи данных в pipe (внутренности)
type Pipeline interface {
	AddPipe(pipe StagePipe, opt *PipelineOpts)
	Start(ctx context.Context) error
	Stop() error
	Input() chan<- Message
	Output() <-chan Message
	Done() chan struct{}
}

// Кол-во воркеров которые будут запускаться на каждый pipe
type PipelineOpts struct {
	MaxWorkers int
}

// Кол-во pipe - набиваем это массив через AddPipe(pipe StagePipe, opt *PipelineOpts)
type WorkersPipeline struct {
	workers []StageWorker
	done    chan struct{}
}

// Добавляем pipe в стек и сразу создаем workers
func (ch *WorkersPipeline) AddPipe(pipe StagePipe, opt *PipelineOpts) {
	//Указываем значение по умолчанию
	if opt == nil {
		opt = &PipelineOpts{MaxWorkers: 1}
	}

	//Создаем каналы для входных и выходных данных
	var input = make(chan Message)
	var output = make(chan Message)

	//Определяем, что входным каналом будет выходной канал ранее созданного pipe
	for _, i := range ch.workers {
		input = i.Output()
	}

	//Создаем для текущего pipe группу работников, кол-во которых определили ранее в PipelineOpts{maxWorkers}
	worker := NewWorkerGroup(opt.MaxWorkers, pipe, input, output)
	// Сохраняем pipe c группой работников в workers
	ch.workers = append(ch.workers, worker)
}

// Канал для входных данных (возвращаем канал последнего pipe из стека WorkersPipeline{workers})
func (ch *WorkersPipeline) Output() <-chan Message {
	sz := len(ch.workers)
	return ch.workers[sz-1].Output()
}

// Канал для выходных данных (возвращаем канал самого первого pipe из стека WorkersPipeline{workers})
func (ch *WorkersPipeline) Input() chan<- Message {
	return ch.workers[0].Input()
}

// Канал для сигнализации всем участником, о том что пора завершить работу
func (ch *WorkersPipeline) Done() chan struct{} {
	return ch.done
}

var ErrConcurrentPipelineEmpty = errors.New("concurrent pipeline empty")

// Запускаем обработку и движение данных в pipelines
func (ch *WorkersPipeline) Start(ctx context.Context) error {
	//Создаем канал для сигнализации о завершении работы всем участникам (горутинам)
	ch.done = make(chan struct{})

	if len(ch.workers) == 0 {
		return ErrConcurrentPipelineEmpty
	}

	//Начинаем обратку данных двигаясь по срезу с pipe-ми и работниками
	for i := 0; i < len(ch.workers); i++ {
		// Получаем группу работников
		g := ch.workers[i]
		//Запускаем паралельную обработку данных (StagePipe{Process(ctx context.Context,step Message) ([]Message, error)}) на уровне одного pipe
		g.Start(ctx, ch.done)
	}

	return nil
}

// Останалвиваем обработку и движение данных в pipeline
func (ch *WorkersPipeline) Stop() error {
	//Крутим срез с работниками и pipe-s
	for _, i := range ch.workers {
		//Закрываем канал входных данных
		//close(i.Input())
		i.WaitStop()
		//Запукаем цепную реакцию, которая закрываем все горутины во всех воркерах
		close(i.Output())
	}

	return nil
}

// Возвращаем указатель на пустую структуру с методами для инициализации pipeline и запуска обработки и движения данных
func NewConcurrentPipeline() Pipeline {
	return &WorkersPipeline{}
}
