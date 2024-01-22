package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
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
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
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
	cfg        *config.Config
	repository repository.Repository
}

func NewGoferHandler(cfg *config.Config, repository repository.Repository) GoferHandler {
	return &goferHandler{
		cfg:        cfg,
		repository: repository,
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

func (gh *goferHandler) Register(c *gin.Context) {
	dataService := service.SetServiceData(c, gh.cfg, gh.repository)

	finished, err := registration.RegistrationRun(dataService)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}
	<-finished
	fmt.Println("STOP REGISTER!")
}

func (gh *goferHandler) Login(c *gin.Context) {
	dataService := service.SetServiceData(c, gh.cfg, gh.repository)

	finished, err := login.LoginRun(dataService, &syncLogin)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP LOGIN!")
}

func (gh *goferHandler) OrdersPost(c *gin.Context) {
	dataService := service.SetServiceData(c, gh.cfg, gh.repository)

	finished, err := accrual.AccrualRun(dataService, &syncOrdersAccrual)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP ADD ORDERS!")
}

func (gh *goferHandler) OrdersGet(c *gin.Context) {
	dataService := service.SetServiceData(c, gh.cfg, gh.repository)

	finished, err := getlistallorders.GetAllListOrtdersRun(dataService)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP LIST ORDERS!")
}

func (gh *goferHandler) Balance(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "Handler.Balance")
	span.SetOperationName("Handler.Mandler.Balance")
	/*span_ := opentracing.SpanFromContext(ctx)
	if span == nil {
		fmt.Println("EMPTY!")
	}*/

	//Получаем TraceID
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		fmt.Println("EMPTY!", sc.TraceID())
		fmt.Println("EMPTY!", sc.SpanID())
	}
	defer span.Finish()

	dataService := service.SetServiceData(c, gh.cfg, gh.repository)

	finished, err := balance.BalanceRun(ctx, dataService)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP GET BALANCE!")
}

func (gh *goferHandler) BalanceWithDraw(c *gin.Context) {

	dataService := service.SetServiceData(c, gh.cfg, gh.repository)

	finished, err := withdraw.WithdrawRun(dataService, &syncWithDraw)

	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP ADD WITHDRAW!")
	// Handle request for version 2 of users route
}

func (gh *goferHandler) WithDrawals(c *gin.Context) {

	dataService := service.SetServiceData(c, gh.cfg, gh.repository)

	finished, err := getlistallwithdrawals.GetAllListWithdrawalsRun(dataService)

	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusBadRequest)
	}

	<-finished
	fmt.Println("STOP ADD WITHDRAW!")
	// Handle request for version 2 of users route
}
