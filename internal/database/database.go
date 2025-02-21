package mydatabase

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gianglt2198/platforms/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

func ProvideDb(cfg *config.Config) *gorm.DB {
	dbOnce.Do(func() {
		logLevel := logger.Silent
		if !cfg.IsProdEnv {
			logLevel = logger.Info
		}

		var err error

		for i := 0; i < 5; i++ {
			db, err = gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
				Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logLevel,
					IgnoreRecordNotFoundError: true,
					ParameterizedQueries:      cfg.IsProdEnv,
					Colorful:                  !cfg.IsProdEnv,
				}),
			})

			if err == nil {
				break
			}
			time.Sleep(time.Second * 5) // retry after 5 seconds
		}

		if err != nil {
			panic(err)
		}

		if cfg.IsProdEnv && cfg.Database.ReplicasConnections != "" {
			_ = db.Use(dbresolver.Register(dbresolver.Config{
				Sources:           []gorm.Dialector{db.Dialector},
				Replicas:          []gorm.Dialector{postgres.Open(cfg.Database.ReplicasConnections)},
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
