package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/accrual"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/balance"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/getlistallorders"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/getlistallwithdrawals"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/login"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/registration"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/withdraw"
	"github.com/gin-gonic/gin"
)

func NewHandlers(ctx context.Context, r *gin.Engine, cfg *config.Config, repository repository.Repository) {

	api := r.Group("/api/user")
	{
		api.POST("/register/", func(c *gin.Context) {

			var sync sync.Mutex

			finished := make(chan struct{})

			registration.RegistrationRun(ctx, c, cfg, repository, finished, &sync)

			<-finished
			fmt.Println("STOP REGISTER!")
		})
		api.POST("/login/", func(c *gin.Context) {

			var sync sync.Mutex

			finished := make(chan struct{})

			login.LoginRun(ctx, c, cfg, repository, finished, &sync)

			<-finished
			fmt.Println("STOP LOGIN!")
		})
		api.POST("/orders/", func(c *gin.Context) {

			var sync sync.Mutex

			finished := make(chan struct{})

			accrual.AccrualRun(ctx, c, cfg, repository, finished, &sync)

			<-finished
			fmt.Println("STOP ADD ORDERS!")
		})
		api.GET("/orders/", func(c *gin.Context) {

			finished := make(chan struct{})

			getlistallorders.GetAllListOrtdersRun(ctx, c, cfg, repository, finished)

			<-finished
			fmt.Println("STOP LIST ORDERS!")
		})
		api.GET("/balance/", func(c *gin.Context) {
			var sync sync.Mutex

			finished := make(chan struct{})

			balance.BalanceRun(ctx, c, cfg, repository, finished, &sync)

			<-finished
			fmt.Println("STOP GET BALANCE!")
		})
		api.POST("/balance/withdraw/", func(c *gin.Context) {

			var sync sync.Mutex

			finished := make(chan struct{})

			withdraw.WithdrawRun(ctx, c, cfg, repository, finished, &sync)

			<-finished
			fmt.Println("STOP ADD WITHDRAW!")
			// Handle request for version 2 of users route
		})
		api.GET("/withdrawals/", func(c *gin.Context) {
			finished := make(chan struct{})

			getlistallwithdrawals.GetAllListWithdrawalsRun(ctx, c, cfg, repository, finished)

			<-finished
			fmt.Println("STOP GET withdrawals!")
		})
	}
}
