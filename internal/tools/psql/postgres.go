package psql

import (
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// DB is the global GORM instance.
// Read and Write operations are automatically routed via DBResolver.
var DB *gorm.DB

// InitPostgres initializes the GORM DB with a Read/Write split.
func InitPostgres(cfg *config.Config) error {
	var err error

	// Connect to Write (Source) Database
	writeDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		cfg.WriteDB.Host, cfg.WriteDB.User, cfg.WriteDB.Password, cfg.WriteDB.NameDB, cfg.WriteDB.Port)

	DB, err = gorm.Open(postgres.Open(writeDsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to write database: %w", err)
	}

	// Connect to Read (Replica) Database via DBResolver
	readDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		cfg.ReadDB.Host, cfg.ReadDB.User, cfg.ReadDB.Password, cfg.ReadDB.NameDB, cfg.ReadDB.Port)

	// Configure DBResolver Plugin
	err = DB.Use(
		dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(readDsn)},
			Policy:   dbresolver.RandomPolicy{}, // Use RandomPolicy for load balancing across multiple replicas
		}).
			SetMaxIdleConns(10).
			SetMaxOpenConns(100).
			SetConnMaxLifetime(time.Hour),
	)

	if err != nil {
		return fmt.Errorf("failed to initialize db resolver plugin: %w", err)
	}

	return nil
}
