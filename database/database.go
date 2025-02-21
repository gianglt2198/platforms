package mydatabase

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gianglt2198/platforms/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

type DBConfig struct {
	IsProdEnv           bool
	Connection          string `json:"connection"`
	ReplicasConnections string `json:"replicas_connections"`
}

var (
	db     *gorm.DB
	dbOnce sync.Once
)

func ProvideDb(cfg *DBConfig) *gorm.DB {
	dbOnce.Do(func() {
		logLevel := logger.Silent
		if !cfg.IsProdEnv {
			logLevel = logger.Info
		}

		db, err := utils.BackoffRetryMechanism(5, func() (*gorm.DB, error) {
			return gorm.Open(postgres.Open(cfg.Connection), &gorm.Config{
				Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logLevel,
					IgnoreRecordNotFoundError: true,
					ParameterizedQueries:      cfg.IsProdEnv,
					Colorful:                  !cfg.IsProdEnv,
				}),
			})
		})

		if err != nil {
			panic(err)
		}

		if cfg.IsProdEnv && cfg.ReplicasConnections != "" {
			_ = db.Use(dbresolver.Register(dbresolver.Config{
				Sources:           []gorm.Dialector{db.Dialector},
				Replicas:          []gorm.Dialector{postgres.Open(cfg.ReplicasConnections)},
				Policy:            dbresolver.RoundRobinPolicy(),
				TraceResolverMode: true,
			}))
		}
	})
	return db
}

func CloseDB() {
	if db != nil {
		sql, _ := db.DB()

		if sql != nil {
			sql.Close()
		}
	}
}
