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

	_ "net/http/pprof"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/config/db"
	configyaml "github.com/dmitryDevGoMid/gofermart/internal/config/yaml"
	"github.com/dmitryDevGoMid/gofermart/internal/handlers"
	"github.com/dmitryDevGoMid/gofermart/internal/migration"
	"github.com/dmitryDevGoMid/gofermart/internal/pb/client"
	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
	"github.com/dmitryDevGoMid/gofermart/internal/pb/server"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/jaeger"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/logger"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/loyalty"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

func Run() {

	////////////////////////////////Трассировка и логирование////////////////
	//Инициализируем логирование и трассировку
	//Извлекаем конф данные
	confLogger, confLogErr := configyaml.ParseConfig()
	if confLogErr != nil {
		log.Fatal(confLogErr)
	}

	//Инициализируем логгер
	appLogger := logger.NewAPILogger(confLogger)
	appLogger.InitLogger()
	appLogger.Info("Start Service API HANDLER")
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, DevelopmentMode: %s",
		confLogger.AppVersion,
		confLogger.Logger.Level,
		confLogger.Server.Development,
	)

	//Инициализируем трайсинг запросов
	tracer, closer, err := jaeger.InitJaeger(confLogger)
	if err != nil {
		appLogger.Fatal("cannot create tracer", err)
	}
	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	////////////////////////////////ТРассировка и логирование////////////////

	ctx, cancel := context.WithCancel(context.Background())
	//Запускаем GrpcServer на 9000 порту
	go server.RunGrpc(ctx, appLogger)

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

	bonusSrvChan := make(chan pb.BonusPlusClient)

	//Запускаем GrpcClient с трассировкой запросов
	go client.RunGrpc(ctx, appLogger, router, bonusSrvChan)

	//Получаем клиента для передачи в стек данных для последующих вызовов
	bonusSrv := <-bonusSrvChan

	close(bonusSrvChan)

	tracing := jaeger.NewTracing(cfg)
	handlersGofermart := handlers.NewGoferHandler(cfg, repository, tracing, bonusSrv)

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
