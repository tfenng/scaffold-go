package di

import (
	"context"
	"database/sql"
	"fmt"

	"scaffold-api/internal/config"
	"scaffold-api/internal/logger"
	"scaffold-api/internal/model"
	"scaffold-api/internal/repository"
	"scaffold-api/internal/server"
	"scaffold-api/internal/service"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewApp(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Supply(cfg),
		fx.Provide(
			logger.New,
			NewDatabase,
			server.NewEcho,
			repository.NewGormUserRepository,
			service.NewUserService,
			server.NewUserHandler,
		),
		fx.Invoke(
			RegisterDatabaseLifecycle,
			server.RegisterHTTPServer,
			server.RegisterDocsRoutes,
			server.RegisterUserRoutes,
			LogStartup,
		),
	)
}

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if cfg.DBAutoMigrate {
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return nil, fmt.Errorf("auto migrate users: %w", err)
		}
	}

	return db, nil
}

func RegisterDatabaseLifecycle(lc fx.Lifecycle, db *gorm.DB, logger zerolog.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	configurePool(sqlDB)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(context.Context) error {
			logger.Info().Msg("closing database connection")
			return sqlDB.Close()
		},
	})

	return nil
}

func LogStartup(logger zerolog.Logger, cfg *config.Config) {
	logger.Info().
		Str("service", cfg.AppName).
		Str("environment", cfg.Environment).
		Msg("application initialized")
}

func configurePool(db *sql.DB) {
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
}
