package models

import "github.com/dmitryDevGoMid/gofermart/internal/pb/pb"

// Структура данных модели
type Plus struct {
	Bonus float32
}

// Функция преобразования к протобуферу
func (u *Plus) ToProtoBuffer() *pb.Plus {
	return &pb.Plus{
		Bonus: u.Bonus,
	}
}

// Преобразуем из протобуфера
func (u *Plus) FromProtoBuffer(plus *pb.Plus) {
	u.Bonus = plus.GetBonus()
}
