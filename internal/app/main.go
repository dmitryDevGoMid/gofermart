package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/config/db"
	configyaml "github.com/dmitryDevGoMid/gofermart/internal/config/yaml"
	"github.com/dmitryDevGoMid/gofermart/internal/handlers"
	"github.com/dmitryDevGoMid/gofermart/internal/migration"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/jaeger"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/logger"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/loyalty"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

// middleware трайсинг
func openTracing(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		span := opentracing.GlobalTracer().StartSpan("apiServer")
		c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))
		log.Error(c.Request.Context(), "Api:=>%v", c.Request.URL)
		c.Next()
	}
}

func Run() {

	////////////////////////////////Трассировка и логирование////////////////
	//Инициализируем логирование и трассировку
	//Извлекаем конф данные
	conf_logger, conf_log_err := configyaml.ParseConfig()
	if conf_log_err != nil {
		log.Fatal(conf_log_err)
	}

	//Инициализируем логгер
	appLogger := logger.NewApiLogger(conf_logger)
	appLogger.InitLogger()
	appLogger.Info("Start Service API HANDLER")
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, DevelopmentMode: %s",
		conf_logger.AppVersion,
		conf_logger.Logger.Level,
		conf_logger.Server.Development,
	)

	//Инициализируем трайсинг запросов
	tracer, closer, err := jaeger.InitJaeger(conf_logger)
	if err != nil {
		appLogger.Fatal("cannot create tracer", err)
	}
	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	/*appLogger.Info("Opentracing connected")

	opts, err := getOpts(appLogger)
	if err != nil {
		appLogger.Fatal("cannot get option for gRPC client", err)
	}
	conn, err := grpc.Dial(authAddr, opts...)
	if err != nil {
		log.Panicln(err)
	}*/

	////////////////////////////////ТРассировка и логирование////////////////

	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.ParseConfig()

	if err != nil {
		fmt.Println("Config", err)
	}

	fmt.Println("ADDRESS_ACRUAL:", cfg.AccrualSystem.AccrualSystem)

	dbConnection := db.NewConnection(cfg)

	dbMigration := migration.NewMigration(dbConnection.DB(), cfg)

	//Инициализируем репозитарий
	repository := repository.NewRepository(dbConnection.DB())

	//Удаляем таблицы из БД
	//dbMigration.RunDrop(ctx)

	//Создаем таблицы в БД
	dbMigration.RunCreate(ctx)

	//Заполняем справочники
	repository.InitCatalogData(ctx)

	router := gin.Default()

	//Указываем мидделвари для трассировки
	router.Use(openTracing(appLogger))

	handlersGofermart := handlers.NewGoferHandler(cfg, repository)

	//Запускаем обработчики запросов http
	handlers.SetHandlers(router, handlersGofermart)

	fmt.Println(cfg.Server.Address)

	//Запускаем запросы к системе лояльности
	go loyalty.Start(ctx, cfg, repository)

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	<-signalChannel
	log.Println("Shutdown Server ...")

	cancel()

	time.Sleep(1 * time.Second)

	dbConnection.Close()

	// Line 51
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}

/*func Run() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.ParseConfig()

	if err != nil {
		fmt.Println("Config", err)
	}

	fmt.Println("ADDRESS_ACRUAL:", cfg.AccrualSystem.AccrualSystem)

	dbConnection := db.NewConnection(cfg)

	dbMigration := migration.NewMigration(dbConnection.DB(), cfg)

	router := gin.Default()

	//Инициализируем репозитарий
	repository := repository.NewRepository(dbConnection.DB())

	//Запускаем обработчики запросов http
	handlers.NewHandlers(ctx, router, cfg, repository)

	fmt.Println(cfg.Server.Address)

	//Удаляем таблицы из БД
	//dbMigration.RunDrop(ctx)

	//Создаем таблицы в БД
	dbMigration.RunCreate(ctx)

	//Заполняем справочники
	repository.InitCatalogData(ctx)

	//Запускаем запросы к системе лояльности
	go loyalty.Start(ctx, cfg, repository)

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	<-signalChannel
	log.Println("Shutdown Server ...")

	cancel()

	time.Sleep(1 * time.Second)

	dbConnection.Close()

	// Line 51
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}*/
