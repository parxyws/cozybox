package main

import (
	"log"

	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/tools/aws"
	"github.com/parxyws/cozybox/internal/tools/psql"
	"github.com/parxyws/cozybox/internal/tools/redis"
)

func main() {
	// 1. Initialize configurations (this will also trigger the validator we just set up)
	cfg, err := config.InitAppConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 2. Initialize PostgreSQL databases (Read and Write)
	if err := psql.InitPostgres(cfg); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	log.Printf("PostgreSQL connections initialized successfully")

	// 3. Initialize Redis
	if err := redis.InitRedis(cfg); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Printf("Redis connection initialized successfully")

	// 4. Initialize AWS (MinIO/S3)
	if err := aws.InitMinio(cfg); err != nil {
		log.Fatalf("Failed to initialize AWS Minio: %v", err)
	}
	log.Printf("AWS Minio client initialized successfully")

	// Ready to use validated configurations and DB global instances!
	log.Printf("Starting application...")
	log.Printf("Server Host: %s", cfg.Server.Host)
	log.Printf("Server Port: %d", cfg.Server.Port)
	log.Printf("Database Name: %s", cfg.WriteDB.NameDB)

	// TODO: Initialize Logger (e.g., Zap/Slog) using cfg.Logger
	// TODO: Start HTTP Server (Gin, Fiber, etc.) on cfg.Server.Port
}
