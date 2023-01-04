package aruna

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rlanhellas/aruna/config"
	"github.com/rlanhellas/aruna/global"
	"github.com/rlanhellas/aruna/httpbridge"
	"github.com/rlanhellas/aruna/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strconv"
	"strings"
)

func setupLogger() {
	var l *zap.Logger
	var err error
	level, err := zap.ParseAtomicLevel(viper.GetString(global.LoggerLevel))
	if err != nil {
		panic(err)
	}

	l, err = zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: viper.GetString(global.LoggerEncoding),
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()

	if err != nil {
		panic(err)
	}

	logger.SetLogger(l.With(zap.String("app", viper.GetString(global.AppName)),
		zap.String("version", viper.GetString(global.AppVer))).Sugar())
}
func setupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("../../../")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
func setupMetrics() {}
func setupHttpServer(routes []*httpbridge.RouteHttp, ctx context.Context) {
	r := gin.Default()

	//process routes
	for _, route := range routes {
		r.Handle(route.Method, route.Path, func(ginctx *gin.Context) {
			httpbridge.HttpHandler(ginctx, ctx, route)
		})
	}

	err := r.Run("0.0.0.0:" + strconv.Itoa(int(config.HttpServerPort())))
	if err != nil {
		panic(err)
	}
}
func setupAuthZAuthN() {}
