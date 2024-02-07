package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
	jaegerTraicing "github.com/dmitryDevGoMid/gofermart/internal/pkg/jaeger"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/accrual"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/balance"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/getlistallorders"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/getlistallwithdrawals"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/login"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/registration"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/withdraw"
	"github.com/gin-gonic/gin"
)

var syncLogin, syncOrdersAccrual, syncWithDraw sync.Mutex

type GoferHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	OrdersPost(c *gin.Context)
	OrdersGet(c *gin.Context)
	Balance(c *gin.Context)
	BalanceWithDraw(c *gin.Context)
	WithDrawals(c *gin.Context)
}

type goferHandler struct {
	tracing    jaegerTraicing.TraicingInterface
	cfg        *config.Config
	repository repository.Repository
	pbCleint   pb.BonusPlusClient
}

func NewGoferHandler(cfg *config.Config, repository repository.Repository, tracing jaegerTraicing.TraicingInterface, pbCleint pb.BonusPlusClient) GoferHandler {
	return &goferHandler{
		tracing:    tracing,
		cfg:        cfg,
		repository: repository,
		pbCleint:   pbCleint,
	}
}

func SetHandlers(r *gin.Engine, gh GoferHandler) {

	api := r.Group("/api/user")
	{
		api.POST("/register/", gh.Register)

		api.POST("/login/", gh.Login)

		api.POST("/orders/", gh.OrdersPost)

		api.GET("/orders/", gh.OrdersGet)

		api.GET("/balance/", gh.Balance)

		api.POST("/balance/withdraw/", gh.BalanceWithDraw)

		api.GET("/withdrawals/", gh.WithDrawals)
	}
}

// Возвраащем линк на стек с данными
func (gh *goferHandler) getServiceData(c *gin.Context) *service.Data {
	return service.SetServiceData(c, gh.cfg, gh.repository, gh.tracing, gh.pbCleint)
}

func (gh *goferHandler) Register(c *gin.Context) {
	span, ctx := gh.tracing.Tracing(c.Request.Context(), "Handler.Register")

	if span != nil {
		defer span.Finish()
	}

	//Инициализируем один раз объект для доступа к репозитарию, конфигу...
	dataService := gh.getServiceData(c)

	finished, err := registration.RegistrationRun(ctx, dataService)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}
	<-finished
	fmt.Println("STOP REGISTER!")
}

func (gh *goferHandler) Login(c *gin.Context) {
	span, ctx := gh.tracing.Tracing(c.Request.Context(), "Handler.Login")

	if span != nil {
		defer span.Finish()
	}

	//Инициализируем один раз объект для доступа к репозитарию, конфигу...
	dataService := gh.getServiceData(c)

	finished, err := login.LoginRun(ctx, dataService, &syncLogin)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP LOGIN!")
}

func (gh *goferHandler) OrdersPost(c *gin.Context) {

	span, ctx := gh.tracing.Tracing(c.Request.Context(), "Handler.OrdersPost")

	if span != nil {
		span.SetOperationName("Handler.Mandler.OrdersPost")
		defer span.Finish()
	}
	//Инициализируем один раз объект для доступа к репозитарию, конфигу...
	dataService := gh.getServiceData(c)

	finished, err := accrual.AccrualRun(ctx, dataService, &syncOrdersAccrual)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP ADD ORDERS!")
}

func (gh *goferHandler) OrdersGet(c *gin.Context) {

	span, ctx := gh.tracing.Tracing(c.Request.Context(), "Handler.OrdersGet")

	if span != nil {
		span.SetOperationName("Handler.Mandler.OrdersGet")
		defer span.Finish()
	}

	//Инициализируем один раз объект для доступа к репозитарию, конфигу...
	dataService := gh.getServiceData(c)

	finished, err := getlistallorders.GetAllListOrtdersRun(ctx, dataService)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP LIST ORDERS!")
}

func (gh *goferHandler) Balance(c *gin.Context) {
	span, ctx := gh.tracing.Tracing(c.Request.Context(), "Handler.Balance")

	if span != nil {
		defer span.Finish()
	}
	//Инициализируем один раз объект для доступа к репозитарию, конфигу...
	dataService := gh.getServiceData(c)

	finished, err := balance.BalanceRun(ctx, dataService)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP GET BALANCE!")
}

func (gh *goferHandler) BalanceWithDraw(c *gin.Context) {
	span, ctx := gh.tracing.Tracing(c.Request.Context(), "Handler.BalanceWithDraw")

	if span != nil {
		span.SetOperationName("Handler.Mandler.BalanceWithDraw")
		defer span.Finish()
	}

	//Инициализируем один раз объект для доступа к репозитарию, конфигу...
	dataService := gh.getServiceData(c)

	finished, err := withdraw.WithdrawRun(ctx, dataService, &syncWithDraw)

	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP ADD WITHDRAW!")
	// Handle request for version 2 of users route
}

func (gh *goferHandler) WithDrawals(c *gin.Context) {

	span, ctx := gh.tracing.Tracing(c.Request.Context(), "Handler.WithDrawals")

	if span != nil {
		span.SetOperationName("Handler.Mandler.WithDrawals")
		defer span.Finish()
	}

	//Инициализируем один раз объект для доступа к репозитарию, конфигу...
	dataService := gh.getServiceData(c)

	finished, err := getlistallwithdrawals.GetAllListWithdrawalsRun(ctx, dataService)

	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP ADD WITHDRAW!")
	// Handle request for version 2 of users route
}
