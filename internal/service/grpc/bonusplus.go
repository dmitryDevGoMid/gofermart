package grpc

import (
	"context"
	"math/rand"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/pb/models"
	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
)

type BonusPlus interface {
}

type BonusPlusService struct {
}

// Возвращаем указатель на сервис, в котором реализуем описанный в proto методы
func NewBonusPlusService() pb.BonusPlusServer {
	return &BonusPlusService{}
}

// Возвращаем дополнительный бонус в зависимости о кол-ва заказов у клиента
func (bp *BonusPlusService) GetBonusPlus(ctx context.Context, req *pb.User) (*pb.Plus, error) {
	user := new(models.User)
	//Преобразуем из protobuf
	user.FromProtoBuffer(req)

	var min float32

	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)

	if user.CountOrders < 10 {
		min = 10
	} else if user.CountOrders < 20 {
		min = 20
	} else if user.CountOrders < 50 {
		min = 30
	} else if user.CountOrders < 100 {
		min = 50
	} else {
		min = 70
	}

	bonus := (rand.Float32() * 10) + min

	plus := new(models.Plus)
	plus.Bonus = bonus

	//Преобразуем в protobuf
	return plus.ToProtoBuffer(), nil
}
