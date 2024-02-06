package models

import (
	"github.com/dmitryDevGoMid/gofermart/internal/pb/pb"
)

// Структура данных модели
type User struct {
	ID          int32
	Name        string
	CountOrders int32
}

// Функция преобразования к протобуферу
func (u *User) ToProtoBuffer() *pb.User {
	return &pb.User{
		Id:          u.ID,
		Name:        u.Name,
		Countorders: u.CountOrders,
	}
}

// Преобразуем из протобуфера
func (u *User) FromProtoBuffer(user *pb.User) {
	u.ID = user.Id
	u.Name = user.GetName()
	u.CountOrders = user.GetCountorders()
}
