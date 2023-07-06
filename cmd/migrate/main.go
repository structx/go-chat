// Package main entrypoint for database migration
package main

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/trevatk/go-pkg/db"
	"github.com/trevatk/go-pkg/logging"
)

func main() {

	fx.New(
		fx.Provide(provideLogger, db.NewSQLite),
		fx.Invoke(registerHooks),
	).Run()
}

func provideLogger() (*zap.Logger, context.Context) {

	l := logging.NewLoggerFromEnv()
	ctx := logging.WithLogger(context.TODO(), l)

	return l.Desugar(), ctx
}

func registerHooks(lc fx.Lifecycle, log *zap.Logger, sdb *sql.DB) error {

	l := log.Sugar()

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {

				l.Info("begin database migration")

				e := db.MigrateSQLite(sdb)
				if e != nil {
					return fmt.Errorf("unable to migrate sqlite db %v", e)
				}

				l.Info("database migration complete")

				return nil
			},
			OnStop: func(ctx context.Context) error {

				e := sdb.Close()
				if e != nil {
					return fmt.Errorf("unable to close database connection %v", e)
				}

				return nil
			},
		},
	)

	return nil
}
