package client

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/logger"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

var (
	portClient     int
	authAddrClient string
)

func init() {
	flag.IntVar(&portClient, "portclient", 9000, "Port to connect to")
	flag.StringVar(&authAddrClient, "authaddrclient", "localhost:9001", "auth service address")
	flag.Parse()
}

func RunGrpc(ctx context.Context, appLogger logger.Logger, route *gin.Engine, bonusService chan pb.BonusPlusClient) {

	appLogger.Info("Opentracing connected")

	opts, err := getOpts(appLogger)
	if err != nil {
		appLogger.Fatal("cannot get option for gRPC client", err)
	}
	conn, err := grpc.Dial(authAddrClient, opts...)
	if err != nil {
		log.Panicln(err)
	}

	bonusPlusClient := pb.NewBonusPlusClient(conn)

	bonusService <- bonusPlusClient

	defer func() {
		fmt.Println("CLOSE GRPC CLIENT")
		conn.Close()
	}()

	route.Use(openTracing(appLogger))

	fmt.Println("gRPC client")

	<-ctx.Done()

	return
}

// Опции для запуска gRPC клиента
func getOpts(logger logger.Logger) ([]grpc.DialOption, error) {
	return []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			middleware.ClientTracing(opentracing.GlobalTracer(), logger),
		)),
	}, nil
}

// middleware трайсинг
func openTracing(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		span := opentracing.GlobalTracer().StartSpan("apiServer")
		c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))
		log.Error(c.Request.Context(), "Api:=>%v", c.Request.URL)
		c.Next()
	}
}
