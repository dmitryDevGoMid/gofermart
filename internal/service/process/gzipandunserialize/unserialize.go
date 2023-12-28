package gzipandunserialize

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/repository"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type UnserializeUser struct{}

func (chain UnserializeUser) Process(result pipeline.Message) ([]pipeline.Message, error) {

	fmt.Println("Processing UnserializeUser")

	data := result.(*service.Data)

	var user repository.User

	err := json.Unmarshal(data.Default.Body, &user)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err
	}

	data.User.User = user
	data.User.UserRequest = user

	return []pipeline.Message{data}, nil
}

type UnserializeLogin struct{}

func (chain UnserializeLogin) Process(result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	var user repository.User

	err := json.Unmarshal(data.Default.Body, &user)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err
	}

	data.User.User = user
	data.User.UserRequest = user

	return []pipeline.Message{data}, nil
}

type UnserializeWithdraw struct{}

func (chain UnserializeWithdraw) Process(result pipeline.Message) ([]pipeline.Message, error) {

	data := result.(*service.Data)

	var withdraw repository.Withdraw

	err := json.Unmarshal(data.Default.Body, &withdraw)

	if err != nil {
		data.Default.ResponseError = func() {
			data.Default.Ctx.Status(http.StatusBadRequest)
		}
		return []pipeline.Message{data}, err
	}

	data.Withdraw.Withdraw = withdraw

	return []pipeline.Message{data}, nil
}
