package db

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/dmitryDevGoMid/gofermart/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var connectPostgres *pgxpool.Pool

// Used during creation of singleton client object
var err error

// Used to execute client creation procedure only once.
var postgresOnce sync.Once

// Интерфейс которыq будет реализован в структуре conn
type Connection interface {
	Close() error
	DB() *pgxpool.Pool
	Ping() error
}

// Структура которая будет возвращать
type conn struct {
	connection *pgxpool.Pool
	cfg        *config.Config
}

// Выполняем соединение с базой данных и возвращаем коннект вместе с ошибкой
func NewConnection(cfg *config.Config) Connection {

	connectPostgres, err := GetPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error opening database: ", err)
	}

	return &conn{connection: connectPostgres, cfg: cfg}
}

// func CloseClientDB() {
func (c *conn) Close() error {
	c.connection.Close() //(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// Реализуем функции на структуре conn для того чтобы она соответствовала интерфейсу Connection
func (c *conn) DB() *pgxpool.Pool {
	return c.connection
}

func (c *conn) Ping() error {
	err = c.connection.Ping(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// Получаем одно соединение для базы данных
func GetPostgresConnection(cfg *config.Config) (*pgxpool.Pool, error) {
	postgresOnce.Do(func() {

		/*connectPostgres, err = pgx.Connect(context.Background(), cfg.DataBase.DatabaseURL)
		if err != nil {
			fmt.Println("Error opening database: ", err)
		}*/

		connectPostgres, err = pgxpool.New(context.Background(), cfg.DataBase.DatabaseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
			os.Exit(1)
		}

		//connectPostgres.Exec(`set search_path='public'`)

		//defer db.Close()

	})

	return connectPostgres, err
}
