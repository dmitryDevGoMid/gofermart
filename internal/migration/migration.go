package migration

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/luna"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var tableList = [...]string{"users", "type_status", "user_accrual", "user_withdraw", "insert_test_data"}
var tableListDrop = [...]string{"user_accrual", "user_withdraw", "users", "type_status"}

type Migration interface {
	Run(ctx context.Context)
}

type migration struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewMigration(db *pgxpool.Pool, cfg *config.Config) *migration {
	return &migration{db: db, cfg: cfg}
}

func (m *migration) RunCreate(ctx context.Context) {
	for _, tableName := range tableList {
		if !m.existsTable(ctx, tableName) {
			m.CreateTable(tableName)
		}
	}
}

func (m *migration) RunDrop(ctx context.Context) {
	for _, tableName := range tableListDrop {
		if m.existsTable(ctx, tableName) {
			m.DropTable(tableName)
		}
	}
}

func (m *migration) existsTable(ctx context.Context, tableName string) bool {
	var n int64
	err := m.db.QueryRow(ctx, "select 1 from information_schema.tables where table_name = $1", tableName).Scan(&n)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false
		}
		return false
	}

	return true
}

/*
Example
`
	CREATE TABLE IF NOT EXISTS TEST(
			UserID int
	)
`
*/

func (m *migration) DropTable(tabelName string) {
	tableCodeDrop := ""

	switch val := tabelName; val {
	case "type_status":
		tableCodeDrop = m.TypeStatusDrop()
	case "user_accrual":
		tableCodeDrop = m.AccrualDrop()
	case "user_withdraw":
		tableCodeDrop = m.WithDrawDrop()
	case "user_balance":
		tableCodeDrop = m.UserBalanceDrop()
	case "users":
		tableCodeDrop = m.UsersDrop()
	default:
		fmt.Println("Not condition case: ", val)
	}

	if tableCodeDrop != "" {
		db, err := m.db.Exec(context.Background(), tableCodeDrop)
		if err != nil {
			fmt.Printf("Error drop table: %v\n", err)
		}

		fmt.Printf("Value of db %v\n", db)
	}
}

func (m *migration) AccrualDrop() string {
	return `drop table user_accrual`
}

func (m *migration) WithDrawDrop() string {
	return `drop table user_withdraw`
}

func (m *migration) UserBalanceDrop() string {
	return `drop table user_balance`
}

func (m *migration) UsersDrop() string {
	return `drop table users`
}

func (m *migration) TypeStatusDrop() string {
	return `drop table type_status`
}

func (m *migration) CreateTable(tabelName string) {

	tableCodeCreate := ""

	switch val := tabelName; val {
	case "type_status":
		tableCodeCreate = m.TypeStatus()
	case "user_accrual":
		tableCodeCreate = m.UserAccrual()
	case "user_withdraw":
		tableCodeCreate = m.UserWithDraw()
	case "users":
		tableCodeCreate = m.Users()
	case "insert_test_data":
		tableCodeCreate = m.SetDataForTest()
	default:
		fmt.Println("Not condition case: ", val)
	}

	if tableCodeCreate != "" {
		db, err := m.db.Exec(context.Background(), tableCodeCreate)

		if err != nil {
			fmt.Printf("Error create table: %v\n", err)
		}
		fmt.Printf("Value of db %v\n", db)
	}
}

func (m *migration) TypeStatus() string {
	return `
	CREATE TABLE IF NOT EXISTS type_status(
		id INT GENERATED ALWAYS AS IDENTITY,
		name VARCHAR(25) NOT NULL,
		PRIMARY KEY(id)
	);
	`
}

func (m *migration) UserAccrual() string {
	return `
	CREATE TABLE IF NOT EXISTS user_accrual(
		id INT GENERATED ALWAYS AS IDENTITY,
		id_user INT NOT NULL,
		id_order NUMERIC NOT NULL,
		accrual decimal(10,2) NOT NULL,
		id_status INT NOT NULL,
		uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(id),
		CONSTRAINT fk_user_accrual
			FOREIGN KEY(id_user) 
				REFERENCES users(id),
		CONSTRAINT fk_status_type
			FOREIGN KEY(id_status) 
				REFERENCES type_status(id)
	);

	CREATE INDEX accrual_id_user ON user_accrual USING btree (id_user);

	`
}

func (m *migration) UserWithDraw() string {
	return `
	CREATE TABLE IF NOT EXISTS user_withdraw(
		id INT GENERATED ALWAYS AS IDENTITY,
		id_user INT NOT NULL,
		id_order NUMERIC NOT NULL,
		sum decimal(10,2) NOT NULL,
		id_status INT NOT NULL,
		processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(id),
		CONSTRAINT fk_user_with_draw
			FOREIGN KEY(id_user) 
				REFERENCES users(id),
		CONSTRAINT fk_status_type
			FOREIGN KEY(id_status) 
				REFERENCES type_status(id)
	);
	CREATE INDEX withdraw_id_user ON user_withdraw USING btree (id_user);
	`
}

func (m *migration) Users() string {
	return `
	CREATE TABLE IF NOT EXISTS users(
		id INT GENERATED ALWAYS AS IDENTITY,
		login VARCHAR(25) NOT NULL,
		password VARCHAR(255) NOT NULL,
		PRIMARY KEY(id)
	);
	`
}

//Отказываемся от таблицы ведения отдельного счета пользователя - добавляем индексы по номеру пользователя
/*func (m *migration) UserBalance() string {
	return `
	CREATE TABLE IF NOT EXISTS user_balance(
		id INT GENERATED ALWAYS AS IDENTITY,
		id_user INT NOT NULL,
		accrual decimal(10,2) NOT NULL,
		withdraw decimal(10,2) NOT NULL,
		PRIMARY KEY(id),
		CONSTRAINT fk_user_balance
			FOREIGN KEY(id_user)
				REFERENCES users(id)
	);
	`
}*/

/*
id INT GENERATED ALWAYS AS IDENTITY,
		id_user INT NOT NULL,
		id_order NUMERIC NOT NULL,
		accrual decimal(10,2) NOT NULL,
		id_status INT NOT NULL,
*/

func (m *migration) SetDataForTest() string {
	if m.cfg.TestDataAdd.Yes == 0 {
		return ""
	}

	insert := ""
	for i := 0; i < 11500; i++ {

		rand.Seed(time.Now().UnixNano())
		min := 1
		max := 5
		id_status := rand.Intn(max-min+1) + min
		accrual := (rand.Float32() * 10) + 10

		order := luna.Generate(10)
		insert = insert + fmt.Sprintf("INSERT INTO user_accrual(id_user,id_order,accrual, id_status) VALUES(1,'%s',%v,%d);\n", order, accrual, id_status)
	}

	return `
	INSERT INTO type_status(name) VALUES('NEW');
	INSERT INTO type_status(name) VALUES('PROCESSING');
	INSERT INTO type_status(name) VALUES('INVALID');
	INSERT INTO type_status(name) VALUES('PROCESSED');
	INSERT INTO type_status(name) VALUES('REGISTERED');

	INSERT INTO users(login,password) VALUES('opsegorsmall@email.ro','$2a$10$FX4hibP9D0Tw0E4EqERQ/eA7mMt5HXFg59F2lHRfS7ZEOIzvppevy');

	` + insert

	//
}
