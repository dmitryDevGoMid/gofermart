package gzipandunserialize

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
	"github.com/gin-gonic/gin"
)

type UnserializeUser struct{}

func (chain UnserializeUser) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {

	fmt.Println("Processing UnserializeUser")

	data := result.(*service.Data)

	var user repository.User

	err := json.Unmarshal(data.Default.Body, &user)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err
	}

	data.User.User = user
	data.User.UserRequest = user

	return []pipeline.Message{data}, nil
}

type UnserializeLogin struct{}

func (chain UnserializeLogin) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	var user repository.User

	err := json.Unmarshal(data.Default.Body, &user)

	if err != nil {
		/*data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}*/
		data.Default.Response = func() {
			data.Default.Ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": string("Bad Request"), // cast it to string before showing
			})
		}
		return []pipeline.Message{data}, err
	}

	data.User.User = user
	data.User.UserRequest = user

	return []pipeline.Message{data}, nil
}

type UnserializeWithdraw struct{}

func (chain UnserializeWithdraw) Process(ctx context.Context, result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	var withdraw repository.Withdraw

	err := json.Unmarshal(data.Default.Body, &withdraw)

	if err != nil {
		data.Default.Response = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err
	}

	data.Withdraw.Withdraw = withdraw

	return []pipeline.Message{data}, nil
}
