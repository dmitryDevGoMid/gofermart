package logger

import (
	config "github.com/dmitryDevGoMid/gofermart/internal/config/yaml"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Структура для реализации интерфейса
type apiLogger struct {
	cfg         *config.Config
	sugarLogger *zap.SugaredLogger
}

// Logger methods interface
type Logger interface {
	InitLogger()
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Printf(template string, args ...interface{})
}

// Перемеррам с уровнями логирования
var loggerLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

// Конструктор для логгера возвращает структура с методами интерфейса
func NewApiLogger(cfg *config.Config) *apiLogger {
	return &apiLogger{cfg: cfg}
}

// Дергаем уровень из карты, если не существует взвращаем zapcore.DebugLevel
func (l *apiLogger) getLoggerLevel(cfg *config.Config) zapcore.Level {
	level, exist := loggerLevelMap[cfg.Logger.Level]
	if !exist {
		return zapcore.DebugLevel
	}

	return level
}

func (l *apiLogger) InitLogger() {
	//Название файла куда пишем логи
	fileName := "zap.log"
	//Получаем уровень логирования
	logLevel := l.getLoggerLevel(l.cfg)

	//logWriter := zapcore.AddSync(os.Stderr)
	logWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename: fileName,
		//MaxSize:   1 << 30, //1G
		MaxSize:   1, //1M
		LocalTime: true,
		Compress:  true,
	})

	var encoderCfg zapcore.EncoderConfig
	if l.cfg.Server.Development {
		//Паникуем и выводим данные
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		//Паникуем и пише в трейс
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	//EncoderConfig устанавливаем параметры для кодировщика
	var encoder zapcore.Encoder

	encoderCfg.LevelKey = "LEVEL"
	encoderCfg.CallerKey = "CALLER"
	encoderCfg.TimeKey = "TIME"
	encoderCfg.NameKey = "NAME"
	encoderCfg.MessageKey = "MESSAGE"

	//Оптимизация вывода в консоль или для оператора в удобном формате JSON
	if l.cfg.Logger.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	//Параметры для кодировщика формат времени
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	//NewCore создает ядро, которое записывает журналы в WriteSyncer.
	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))

	//New создает новый регистратор из предоставленных zapcore.Core и Options. Если переданный zapcore.Core равен нулю, он возвращается к использованию бездействующей реализации.
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	l.sugarLogger = logger.Sugar()
	if err := l.sugarLogger.Sync(); err != nil {
		l.sugarLogger.Error(err)
	}
}

// Logger methods

func (l *apiLogger) Debug(args ...interface{}) {
	l.sugarLogger.Debug(args...)
}

func (l *apiLogger) Debugf(template string, args ...interface{}) {
	l.sugarLogger.Debugf(template, args...)
}

func (l *apiLogger) Info(args ...interface{}) {
	l.sugarLogger.Info(args...)
}

func (l *apiLogger) Infof(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

func (l *apiLogger) Printf(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

func (l *apiLogger) Warn(args ...interface{}) {
	l.sugarLogger.Warn(args...)
}

func (l *apiLogger) Warnf(template string, args ...interface{}) {
	l.sugarLogger.Warnf(template, args...)
}

func (l *apiLogger) Error(args ...interface{}) {
	l.sugarLogger.Error(args...)
}

func (l *apiLogger) Errorf(template string, args ...interface{}) {
	l.sugarLogger.Errorf(template, args...)
}

func (l *apiLogger) DPanic(args ...interface{}) {
	l.sugarLogger.DPanic(args...)
}

func (l *apiLogger) DPanicf(template string, args ...interface{}) {
	l.sugarLogger.DPanicf(template, args...)
}

func (l *apiLogger) Panic(args ...interface{}) {
	l.sugarLogger.Panic(args...)
}

func (l *apiLogger) Panicf(template string, args ...interface{}) {
	l.sugarLogger.Panicf(template, args...)
}

func (l *apiLogger) Fatal(args ...interface{}) {
	l.sugarLogger.Fatal(args...)
}

func (l *apiLogger) Fatalf(template string, args ...interface{}) {
	l.sugarLogger.Fatalf(template, args...)
}
