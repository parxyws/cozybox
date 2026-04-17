package main

import (
	"fmt"
	"log"

	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/core"
	"github.com/parxyws/cozybox/internal/tools/aws"
	"github.com/parxyws/cozybox/internal/tools/mail"
	"github.com/parxyws/cozybox/internal/tools/psql"
	"github.com/parxyws/cozybox/internal/tools/redis"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.InitAppConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	fmt.Println(cfg)

	db, err := psql.InitPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}

	authRedis, err := redis.InitAuthRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Auth Redis: %v", err)
	}

	cacheRedis, err := redis.InitCacheRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Cache Redis: %v", err)
	}

	limiterRedis, err := redis.InitLimiterRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Limiter Redis: %v", err)
	}

	minio, err := aws.InitMinio(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize AWS Minio: %v", err)
	}

	gomail := mail.NewGoMailDialer(cfg)

	app := core.NewApp(&core.AppConfig{
		Cfg:          cfg,
		Db:           db,
		Minio:        minio,
		AuthRedis:    authRedis,
		CacheRedis:   cacheRedis,
		LimiterRedis: limiterRedis,
		Logger:       logrus.New(),
		Mail:         gomail,
	})

	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

}
