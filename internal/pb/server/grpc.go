package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/logger"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/middleware"
	grpcservice "github.com/dmitryDevGoMid/gofermart/internal/service/grpc"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	local bool
	port  int
)

func init() {
	flag.IntVar(&port, "port", 9001, "Port to listen on authentification service")
	flag.BoolVar(&local, "local", true, "run service local")
	flag.Parse()
}
func RunGrpc(ctx context.Context, appLogger logger.Logger) {

	bonusService := grpcservice.NewBonusPlusService()

	//Устанавливаем прослушивание tcp на порту
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	//Создаем grpc сервер
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			middleware.ServerTracing(opentracing.GlobalTracer(), appLogger), //jaeger
		),
	))

	defer func() {
		fmt.Println("CLOSE GRPC SERVER")
		//Закрываем tcp
		lis.Close()
		//Корректно завершаем работу сервера
		grpcServer.GracefulStop()
	}()

	// Регистрируем в protobuf вервер и сервис
	pb.RegisterBonusPlusServer(grpcServer, bonusService)

	reflection.Register(grpcServer)

	log.Printf("Starting service running on %d\n", port)

	var errGrpc error

	go func(errGrpc *error) {
		var err error
		err = grpcServer.Serve(lis)
		errGrpc = &err
	}(&errGrpc)

	if errGrpc != nil {
		appLogger.Error(context.Background(), "[rpcServer] 开始监听%v错误 %v", fmt.Sprintf(":%d", port), errGrpc)
	}

	//Жде завершения через контекст
	<-ctx.Done()

	return

}
