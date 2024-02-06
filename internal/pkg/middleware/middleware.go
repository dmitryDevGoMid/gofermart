/*
Простой ман: https://dev.to/davidsbond/golang-creating-grpc-interceptors-5el5
*/
package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/logger"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

/**
Описание итерфейса который используется для распространения трассировки
https://github.com/opentracing/opentracing-go/blob/master/propagation.go#L106
*/

//var carrier map[string]string

type MDCarrier struct {
	metadata.MD
}

func (m MDCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, strs := range m.MD {
		for _, v := range strs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m MDCarrier) Set(key, val string) {
	m.MD[key] = append(m.MD[key], val)
}

// ClientInterceptor унарный перехватчик grpc.UnaryClientInterceptor
func ClientTracing(tracer opentracing.Tracer, log logger.Logger) grpc.UnaryClientInterceptor {
	//Функция, которая соответствует сигнатуре grpc.UnaryClientInterceptor
	/** ctx context.Context— Это контекст запроса, который будет использоваться в основном для таймаутов. Его также можно использовать для добавления/чтения метаданных запроса.
	*** method string— Имя вызываемого метода RPC.
	*** req interface{}- Экземпляр запроса, это interface{}отражение, используемое для маршалинга.
	*** reply interface{}- Экземпляр ответа работает так же, как и reqпараметр.
	*** cc *grpc.ClientConn- Базовое соединение клиента с сервером.
	*** invoker grpc.UnaryInvoker- Метод вызова RPC. Как и в случае с промежуточным программным обеспечением HTTP , где вы вызываете ServeHTTP, его необходимо вызвать для выполнения вызова RPC.
	*** opts ...grpc.CallOption— grpc.CallOptionЭкземпляры, используемые для настройки вызова gRPC.
	 */
	return func(ctx context.Context, method string, request, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//Отношения между вызывающим сервером и клиентом службы
		var parentCtx opentracing.SpanContext
		//carrier := make(map[string]string)
		parentSpan := opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			parentCtx = parentSpan.Context()
		}
		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(parentCtx),
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC Client"},
			ext.SpanKindRPCClient,
		)

		defer span.Finish()
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		err := tracer.Inject(
			span.Context(),
			opentracing.TextMap,
			MDCarrier{md},
			//opentracing.TextMapCarrier(carrier),
		)

		if err != nil {
			log.Error(ctx, "ClientTracing inject span error :%v", err.Error())
		}

		///SiteCode
		siteCode := fmt.Sprintf("%v", ctx.Value("SiteCode"))
		if len(siteCode) < 1 || strings.Contains(siteCode, "nil") {
			siteCode = "001"
		}
		md.Set("SiteCode", siteCode)
		//
		newCtx := metadata.NewOutgoingContext(ctx, md)
		err = invoker(newCtx, method, request, reply, cc, opts...)

		if err != nil {
			log.Error(ctx, "ClientTracing call error : %v", err.Error())
		}
		return err
	}

}

// ServerInterceptor Server - унарный серверный перехватчик
func ServerTracing(tracer opentracing.Tracer, log logger.Logger) grpc.UnaryServerInterceptor {
	//сигнатура функции, она определяется как grpc.UnaryServerInterceptor
	/**
	*** ctx context.Context— Это контекст запроса, который будет использоваться в основном для таймаутов. Его также можно использовать для добавления/чтения метаданных запроса.
	*** req interface{}- Входящий запрос
	*** info *grpc.UnaryServerInfo- Информация о сервере gRPC, обрабатывающем запрос.
	*** handler grpc.UnaryHandler— Обработчик входящего запроса. Вам нужно будет его вызвать, иначе вы не получите ответ клиенту.
	 */
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		spanContext, err := tracer.Extract(
			opentracing.TextMap,
			MDCarrier{md},
		)

		if err != nil && err != opentracing.ErrSpanContextNotFound {
			log.Error(ctx, "ServerInterceptor extract from metadata err: %v", err)
		} else {
			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "(gRPC Server)"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()

			ctx = opentracing.ContextWithSpan(ctx, span)
		}

		return handler(ctx, req)
	}

}
