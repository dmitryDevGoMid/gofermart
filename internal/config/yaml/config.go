package yaml

import (
	"os"
	"time"

	"log"

	"github.com/spf13/viper"
)

type Logger struct {
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}

// Jaeger config
type Jaeger struct {
	Host        string
	ServiceName string
	LogSpans    bool
}

// Server config
type Server struct {
	Port              string
	Development       bool
	Timeout           time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxConnectionIdle time.Duration
	MaxConnectionAge  time.Duration
	//Kafka             Kafka
}

type Config struct {
	AppVersion string
	Logger     Logger
	Server     Server
	Jaeger     Jaeger
}

// Подтягиваем файл конфигурации
func exportConfig() error {
	//Устанавливаем тип конфига
	viper.SetConfigType("yaml")

	//Добавляем путь к конфигу
	viper.AddConfigPath("./internal/config/yaml")
	//Проверяем пременную окружения
	if os.Getenv("MODE") == "DOCKER" {
		viper.SetConfigName("config-docker.yml")
	} else {
		viper.SetConfigName("config.yaml")
	}
	//viper.SetConfigFile("./config/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// Разбираем конфигурацию по структурам
func ParseConfig() (*Config, error) {
	if err := exportConfig(); err != nil {
		return nil, err
	}

	//Обьявляем перменную типа  Config
	var c Config
	//Передаем указатель в функцию Unmarshal viper
	err := viper.Unmarshal(&c)

	if err != nil {
		log.Printf("decode configuration error: %v", err)
		return nil, err
	}

	return &c, nil
}
