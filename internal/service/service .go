/*
   Набор структур с названияеми основных процессов (регистрация, аутентификация, пополнение, списание ...)
   Структуры содержат переменные, которые заполняются по мере продвижения по pipilines процессу
*/

package service

import (
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
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
	HashPassword string
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
