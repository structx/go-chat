// Package main service entrypoint
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/trevatk/go-pkg/db"
	"github.com/trevatk/go-pkg/logging"

	"github.com/trevatk/go-chat/internal/domain"
	"github.com/trevatk/go-chat/internal/port"
	"github.com/trevatk/go-chat/internal/port/middleware"
	pb "github.com/trevatk/go-chat/proto/messenger/v1"
)

func main() {

	fx.New(
		fx.Provide(context.TODO),
		fx.Provide(provideLogger),
		fx.Provide(db.NewSQLite),
		fx.Provide(domain.NewBundle),
		fx.Provide(port.NewHTTPServer),
		fx.Provide(middleware.NewAuthenticator),
		fx.Provide(fx.Annotate(port.NewRouter, fx.As(new(http.Handler)))),
		fx.Provide(port.NewGrpcServer),
		fx.Invoke(registerHooks),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	).Run()
}

func provideLogger(ctx context.Context) (*zap.SugaredLogger, context.Context) {

	l := logging.NewLoggerFromEnv()
	ctx = logging.WithLogger(ctx, l)

	return l, ctx
}

func registerHooks(lc fx.Lifecycle, log *zap.SugaredLogger, handler http.Handler, gSrv *port.GrpcServer, sqlite *sql.DB) error {

	p1 := os.Getenv("HTTP_SERVER_PORT")
	if p1 == "" {
		return errors.New("$HTTP_SERVER_PORT is unset")
	}

	s1 := &http.Server{
		Addr:         ":" + p1,
		Handler:      handler,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 15,
	}

	p2 := os.Getenv("GRPC_SERVER_PORT")
	if p2 == "" {
		return errors.New("$GRPC_SERVER_PORT is unset")
	}

	s2 := grpc.NewServer()
	pb.RegisterMessengerServiceServer(s2, gSrv)

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {

				log.Info("execute database migration")

				e := db.MigrateSQLite(sqlite)
				if e != nil {
					return fmt.Errorf("failed to execute database migration %v", e)
				}

				log.Infof("start http server http://localhost:%s", p1)

				go func() {
					if e := s1.ListenAndServe(); e != nil {
						log.Fatalf("failed to start http server %v", e)
					}
				}()

				li, e := net.Listen("tcp", ":"+p2)
				if e != nil {
					return fmt.Errorf("unable to create network listener %v", e)
				}

				log.Infof("start gRPC server localhost:%s", p2)

				go func() {
					if e := s2.Serve(li); e != nil {
						log.Fatalf("failed to start gRPC server %v", e)
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {

				var e error

				log.Info("close database connection")

				e = sqlite.Close()
				if e != nil {
					log.Errorf("failed to close database connection %v", e)
				}

				log.Info("shutdown http server")

				e = s1.Close()
				if e != nil && !errors.Is(e, http.ErrServerClosed) {
					log.Errorf("failed to shutdown http server %v", e)
				}

				log.Info("shutdown gRPC server")
				s2.GracefulStop()

				// redudant logging
				return e
			},
		},
	)

	return nil
}
