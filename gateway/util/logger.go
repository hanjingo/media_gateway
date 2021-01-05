package util

import (
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logConf *LogConfig
var logConfOnce = new(sync.Once)
var logger *zap.SugaredLogger
var loggerOnce = new(sync.Once)

// writer type
const (
	WriterTypStdou string = "stdout"
	WriterTypFile  string = "file"
)

// encoder type
const (
	EncoderTypConsole string = "console"
	EncoderTypJson    string = "json"
)

type LogConfig struct {
	WriterTypes []string `json:"writer_types"` // stdout, file
	Lvl         string   `json:"lvl"`          // debug, info, warn, error, dpanic, panic, fatal
	Encoding    string   `json:"encoding"`     // console, json
	Filename    string   `json:"filename"`
	MaxSizeMB   int      `json:"max_size_mb"`
	MaxAgeDay   int      `json:"max_age_day"`
	MaxBackup   int      `json:"max_backup"`
	IsLocalTime bool     `json:"is_local_time"`
	IsCompress  bool     `json:"is_compress"`
}

func (c *LogConfig) Check() {
	// todo
}

func Log() *zap.SugaredLogger {
	loggerOnce.Do(func() {
		logger = NewSugaredLogger(LogConf())
	})
	return logger
}

func LogConf() *LogConfig {
	logConfOnce.Do(func() {
		logConf = DefaultLogConfig()
	})
	return logConf
}

func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		WriterTypes: []string{"stdout"},
		Lvl:         "debug",
		Encoding:    "json",
		Filename:    "default.log",
		MaxSizeMB:   2,
		MaxAgeDay:   1,
		MaxBackup:   7,
		IsCompress:  false,
		IsLocalTime: true,
	}
}

func NewSugaredLogger(conf *LogConfig) *zap.SugaredLogger {
	// 打印机
	var writer zapcore.WriteSyncer
	writers := []zapcore.WriteSyncer{}
	for _, typ := range LogConf().WriterTypes {
		switch typ {
		case WriterTypStdou:
			writers = append(writers, os.Stdout)
		case WriterTypFile:
			writers = append(writers, zapcore.AddSync(&lumberjack.Logger{
				Filename:   LogConf().Filename,
				MaxSize:    LogConf().MaxSizeMB,
				MaxAge:     LogConf().MaxAgeDay,
				MaxBackups: LogConf().MaxBackup,
				LocalTime:  LogConf().IsLocalTime,
				Compress:   LogConf().IsCompress,
			}))
		}
	}
	writer = zapcore.NewMultiWriteSyncer(writers...)

	// 编码器
	var encoder zapcore.Encoder
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encCfg.EncodeCaller = zapcore.ShortCallerEncoder
	switch LogConf().Encoding {
	case EncoderTypConsole:
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // 开启颜色
		encoder = zapcore.NewConsoleEncoder(encCfg)
	case EncoderTypJson:
		encoder = zapcore.NewJSONEncoder(encCfg)
	default:
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	// 最小等级
	var lvl zapcore.Level
	if err := lvl.Set(LogConf().Lvl); err != nil {
		lvl = zapcore.DebugLevel
	}

	core := zapcore.NewCore(encoder, writer, lvl)
	return zap.New(core, zap.AddCaller()).Sugar()
}
