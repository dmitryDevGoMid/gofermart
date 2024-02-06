package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof" // подключаем пакет pprof

	"github.com/dmitryDevGoMid/gofermart/internal/app"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	go app.Run()
	srvProf := &http.Server{
		Addr: ":8888",
	}
	go func() {
		// service connections
		if err := srvProf.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	<-signalChannel
	log.Println("Shutdown Server ...")

	cancel()

	time.Sleep(1 * time.Second)

	if err := srvProf.Shutdown(ctx); err != nil {
		log.Fatal("Server srvProf forced to shutdown: ", err)
	}
}
