/*
   Набор структур с названияеми основных процессов (регистрация, аутентификация, пополнение, списание ...)
   Структуры содержат переменные, которые заполняются по мере продвижения по pipilines процессу
*/

package service

import (
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/jaeger"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/gin-gonic/gin"
)

type Data struct {
	Default        Default
	Accrual        Accrual
	Authentication Authentication
	Balance        Balance
	Withdraw       Withdraw
	Password       Password
	User           User
	Claims         Claims
	Loyalty        Loyalty
}

func (current *Data) GetNewService() *Data {
	data := &Data{}

	//Устанавливаем новый канал для стека данных
	finished := make(chan struct{})

	// Устанавливаем контекст запроса gin
	data.Default.Ctx = current.Default.Ctx
	// Устанавливаем конфигурационные данные
	data.Default.Cfg = current.Default.Cfg
	// Устанавливаем канал завершения процесса
	data.Default.Finished = finished
	// Устанавливаем репозитарий для данных из базы
	data.Default.Repository = current.Default.Repository
	// Трассировка
	data.Default.Tracing = current.Default.Tracing
	// Клиент Grpc для запросов к серверу
	data.Default.PbCleint = current.Default.PbCleint

	return data
}

func SetServiceData(c *gin.Context, cfg *config.Config, rep repository.Repository, tracing jaeger.TraicingInterface, pbCleint pb.BonusPlusClient) *Data {
	//func SetServiceData(c context.Context, cfg *config.Config, rep repository.Repository) *Data {
	data := &Data{}

	// Устанавливаем контекст запроса gin
	data.Default.Ctx = c
	// Устанавливаем конфигурационные данные
	data.Default.Cfg = cfg
	// Устанавливаем репозитарий для данных из базы
	data.Default.Repository = rep
	// Трассировка
	data.Default.Tracing = tracing
	// Клиент Grpc для запросов к серверу
	data.Default.PbCleint = pbCleint

	return data
}

type Accrual struct {
	Accrual     repository.Accrual
	Accruals    *[]repository.Accrual
	AccrualList *[]repository.AccrualList
}

type Balance struct {
	Balance         repository.Balance
	ResponseBalance repository.ResponseBalance
}

type Withdraw struct {
	Withdraw     repository.Withdraw
	WithdrawList *[]repository.WithdrawList
}

type User struct {
	User         repository.User
	UserRequest  repository.User
	PbUser       *pb.User
	HashPassword string
	BonusPlus    float32
}

type Claims struct {
	Username string
}

type Authentication struct {
}

type Password struct {
	HashPassword  string
	Password      string
	CheckPassword bool
}

type Default struct {
	Repository    repository.Repository
	Ctx           *gin.Context
	Cfg           *config.Config
	Response      func()
	ResponseError func()
	Finished      chan struct{}
	Body          []byte
	Tracing       jaeger.TraicingInterface
	PbCleint      pb.BonusPlusClient
}

type Loyalty struct {
	Accruals *[]repository.Accrual
	//Один платеж для одной отправки
	Accrual                repository.Accrual
	Response               []byte
	ResponseLoyaltyService ResponseLoyaltyService
	Loyalty                *time.Ticker
}

type ResponseLoyaltyService struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}
