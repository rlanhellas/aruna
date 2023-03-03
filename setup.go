package aruna

import (
	"context"
	"fmt"
	"gorm.io/gorm/schema"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rlanhellas/aruna/config"
	"github.com/rlanhellas/aruna/db"
	"github.com/rlanhellas/aruna/global"
	"github.com/rlanhellas/aruna/httpbridge"
	"github.com/rlanhellas/aruna/logger"
	"github.com/rlanhellas/aruna/security"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	loggergorm "gorm.io/gorm/logger"
)

func setupLogger() {
	var l *zap.Logger
	var err error
	level, err := zap.ParseAtomicLevel(config.LoggerLevel())
	if err != nil {
		panic(err)
	}

	outputPaths := []string{"stdout"}
	if config.LoggerPath() != "" {
		outputPaths = append(outputPaths, config.LoggerPath())
	}

	errOutputPaths := []string{"stderr"}
	if config.LoggerPath() != "" {
		errOutputPaths = append(errOutputPaths, config.LoggerPath())
	}

	l, err = zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: config.LoggerEncoding(),
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
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errOutputPaths,
	}.Build()

	if err != nil {
		panic(err)
	}

	logger.SetLogger(l.With(zap.String("app", config.AppName()),
		zap.String("version", config.AppVer())).Sugar())
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
func setupHttpServer(routesGroup []*httpbridge.RouteGroupHttp, ctx context.Context) {
	r := gin.Default()

	for _, group := range routesGroup {
		g := r.Group(group.Path)
		if group.Authenticated && config.SecurityEnabled() {
			g.Use(func(c *gin.Context) {
				authHeader := c.GetHeader("Authorization")
				if authHeader == "" {
					c.AbortWithStatus(http.StatusForbidden)
				}

				if !security.ValidateJwt(ctx, authHeader) {
					c.AbortWithStatus(http.StatusForbidden)
				}
			})
		}

		mapRoutes := make(map[string]*httpbridge.RouteHttp, len(group.Routes))
		for _, route := range group.Routes {
			mapRoutes[fmt.Sprintf("%s%s@%s", group.Path, route.Path, route.Method)] = route
			g.Handle(route.Method, route.Path, func(ginctx *gin.Context) {
				if innerRoute, ok := mapRoutes[fmt.Sprintf("%s@%s", ginctx.FullPath(), ginctx.Request.Method)]; ok {
					httpbridge.HttpHandler(ginctx, ctx, innerRoute)
				} else {
					ginctx.JSON(http.StatusInternalServerError, "Route not mapped")
				}
			})
		}
	}

	r.GET("/doc/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	err := r.Run("0.0.0.0:" + strconv.Itoa(config.HttpServerPort()))
	if err != nil {
		panic(err)
	}
}
func setupDB(ctx context.Context, migrateTables []any) {

	gormLogLevel := loggergorm.Default.LogMode(loggergorm.Silent)
	if config.DbShowSQL() {
		gormLogLevel = loggergorm.Default.LogMode(loggergorm.Info)
	}

	schemaNameStrategy := schema.NamingStrategy{}
	if config.DbUseSchema() {
		projectSchema := config.DbSchema() + "."
		schemaNameStrategy = schema.NamingStrategy{
			TablePrefix:   projectSchema,
			SingularTable: false,
		}
		logger.Debug(ctx, "connection for schema: %s", projectSchema)
	}

	switch config.DbType() {
	case global.PostgresDBType:
		clientdb, err := gorm.Open(postgres.Open(config.DbConnectionString()), &gorm.Config{
			Logger:         gormLogLevel,
			NamingStrategy: schemaNameStrategy,
		})

		if err != nil {
			panic(err)
		}

		if config.DbSchema() != "" {
			clientdb.Exec(fmt.Sprintf("set search_path='%s'", config.DbSchema()))
		}

		if migrateTables != nil {
			for _, mt := range migrateTables {
				logger.Debug(ctx, "migrating table %s", reflect.TypeOf(mt).String())
				err := clientdb.AutoMigrate(mt)
				if err != nil {
					panic(err)
				}
			}
		}

		db.SetClient(clientdb)
	default:
		panic(fmt.Sprintf("unsupported db type %s", config.DbType()))
	}
}

func setupAuthZAuthN() {}
