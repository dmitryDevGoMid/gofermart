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
	"github.com/dmitryDevGoMid/gofermart/internal/handlers"
	"github.com/dmitryDevGoMid/gofermart/internal/migration"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service/process/loyalty"
	"github.com/gin-gonic/gin"
)

func Run() {
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
}
