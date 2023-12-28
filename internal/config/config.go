package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type AccrualSystem struct {
	AccrualSystem string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type TickerTime struct {
	TickTack int
}

type TestDataAdd struct {
	Yes int
}

type HashSHA256 struct {
	Key string `env:"KEY,omitempty"`
}

type DataBase struct {
	DatabaseURL string `env:"DATABASE_URI"`
}

type File struct {
	Restore         bool   `env:"RESTORE"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
}

type Gzip struct {
	Enable bool `env:"GZIP"`
}

type Serializer struct {
	SerType string `env:"SER_TYPE"`
}

type Logger struct {
	Encoding string `env:"LOG_ENCODING"`
	Level    string `env:"LOG_LEVEL"`
}

type Metrics struct {
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
}

type Server struct {
	Address string `env:"RUN_ADDRESS"`
}

type Config struct {
	Server        Server
	Metrics       Metrics
	Logger        Logger
	Serializer    Serializer
	Gzip          Gzip
	File          File
	DataBase      DataBase
	HashSHA256    HashSHA256
	TestDataAdd   TestDataAdd
	TickerTime    TickerTime
	AccrualSystem AccrualSystem
}

var (
	address string
	//reportInterval int
	pollInterval   int
	loggerEncoding string
	loggerLevel    string
	serializeType  string
	enableGzip     bool

	restoreFile       bool
	storeIntervalFile int
	fileStoragePath   string

	databaseURL string

	keySHA256 string
	testdata  int

	accrualSystemAdress string
)

func init() {
	flag.IntVar(&testdata, "td", 0, "add test data to table")

	flag.StringVar(&address, "a", "localhost:8000", "location http server")
	//flag.IntVar(&reportInterval, "r", 400, "interval for run metrics")
	flag.IntVar(&pollInterval, "p", 200, "interval for run metrics")

	//	Logger
	flag.StringVar(&loggerEncoding, "logen", "console", "set logger config encoding")
	flag.StringVar(&loggerLevel, "loglv", "InfoLevel", "set logger config level")

	//Serialize Type
	flag.StringVar(&serializeType, "sertype", "json", "set logger config encoding")

	//Serialize Type
	flag.BoolVar(&enableGzip, "gzip", false, "set gzip for agent and server")

	//File
	flag.BoolVar(&restoreFile, "r", true, "restore file")
	//flag.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "path file")
	flag.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "path file")
	flag.IntVar(&storeIntervalFile, "i", 300, "store interval file")

	//Connection Database
	/*
			  - POSTGRES_PASSWORD=M45fgMetr
		      - POSTGRES_USER=manager
		      - POSTGRES_DB=metrics
	*/
	flag.StringVar(&databaseURL, "d", "postgres://manager:M45fgMetr@localhost:5432/gofermart?sslmode=disable", "database url for conection postgress")

	//sha 256 key
	flag.StringVar(&keySHA256, "k", "", "set key for calc SHA256")

	//Flag adres accrual system
	flag.StringVar(&accrualSystemAdress, "-r", "http://localhost:8080", "set adress for send loaylte accrual")
}

// Разбираем конфигурацию по структурам
func ParseConfig() (*Config, error) {
	flag.Parse()

	var config Config

	config.Metrics.PollInterval = pollInterval
	//config.Metrics.ReportInterval = reportInterval

	config.Server.Address = address

	config.Logger.Encoding = loggerEncoding
	config.Logger.Level = loggerLevel

	config.Serializer.SerType = serializeType

	config.Gzip.Enable = enableGzip

	config.File.FileStoragePath = fileStoragePath
	config.File.Restore = restoreFile
	config.File.StoreInterval = storeIntervalFile

	config.DataBase.DatabaseURL = databaseURL

	config.HashSHA256.Key = keySHA256

	config.TestDataAdd.Yes = testdata
	config.TickerTime.TickTack = 3

	config.AccrualSystem.AccrualSystem = accrualSystemAdress

	//Init by environment variables
	env.Parse(&config.Metrics)
	env.Parse(&config.Server)
	env.Parse(&config.Logger)
	env.Parse(&config.Serializer)
	env.Parse(&config.Gzip)
	env.Parse(&config.File)
	env.Parse(&config.DataBase)
	env.Parse(&config.HashSHA256)
	env.Parse(&config.AccrualSystem)

	return &config, nil
}
