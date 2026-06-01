package app

import (
	"context"
	"errors"
	"net/http"
	"sekai/internal/adapters/inspector"
	"sekai/internal/adapters/postgres"
	"sekai/internal/api"
	"sekai/internal/api/middlewares"
	"sekai/internal/config"
	scanusecase "sekai/internal/usecase/scan"
	"time"

	"github.com/76Parker/golib/loglib"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var ConfigModule = fx.Options(
	fx.Provide(config.Load),
)

var LoggerModule = fx.Options(
	fx.Provide(newLogger),
)

var PostgresModule = fx.Options(
	fx.Provide(newPostgresPool),
	fx.Provide(fx.Annotate(
		postgres.NewScanRepository,
		fx.As(new(scanusecase.Repository)),
	)),
	fx.Provide(fx.Annotate(
		postgres.NewUserRepository,
		fx.As(new(middlewares.UserRepository)),
	)),
)

var AuthModule = fx.Options(
	fx.Provide(newKeyfunc),
)

var ScanModule = fx.Options(
	fx.Provide(newArtifactInspector),
	fx.Provide(fx.Annotate(
		newScanService,
		fx.As(new(api.ScanService)),
	)),
	fx.Provide(api.NewScanHandler),
)

var HTTPModule = fx.Options(
	fx.Provide(newRouterConfig),
	fx.Provide(fx.Annotate(
		api.NewRouter,
		fx.As(new(http.Handler)),
	)),
	fx.Provide(newHTTPServer),
	fx.Invoke(registerHTTPServer),
)

func New() *fx.App {
	return fx.New(
		fx.NopLogger,
		ConfigModule,
		LoggerModule,
		PostgresModule,
		AuthModule,
		ScanModule,
		HTTPModule,
	)
}

func newLogger(lc fx.Lifecycle, cfg config.Config) (loglib.Logger, error) {
	logger, err := loglib.NewSlog(cfg.Log)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return logger.Close()
		},
	})

	return logger, nil
}

func newPostgresPool(lc fx.Lifecycle, cfg config.Config, logger loglib.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.Postgres.DatabaseURL)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return err
			}
			logger.Info("postgres connection established")
			if err := postgres.Migrate(ctx, pool, cfg.Postgres.MigrationsDir); err != nil {
				return err
			}
			logger.Info("postgres migrations applied", "dir", cfg.Postgres.MigrationsDir)
			return nil
		},
		OnStop: func(context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

func newKeyfunc(lc fx.Lifecycle, cfg config.Config) (keyfunc.Keyfunc, error) {
	if !cfg.Auth.Enabled {
		return keyfunc.NewJWKSetJSON([]byte(`{"keys":[]}`))
	}

	ctx, cancel := context.WithCancel(context.Background())
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{cfg.Auth.JWKSURL})
	if err != nil {
		cancel()
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})

	return jwks, nil
}

func newArtifactInspector(cfg config.Config) scanusecase.ArtifactInspector {
	return inspector.NewZipInspector(cfg.Inspector.MaxUncompressedSizeBytes)
}

func newScanService(
	artifactInspector scanusecase.ArtifactInspector,
	repository scanusecase.Repository,
) *scanusecase.Service {
	return scanusecase.NewService(artifactInspector, repository, nil)
}

func newRouterConfig(cfg config.Config) api.RouterConfig {
	return api.RouterConfig{
		AuthEnabled: cfg.Auth.Enabled,
		AuthIssuer:  cfg.Auth.Issuer,
	}
}

func newHTTPServer(cfg config.Config, router http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}
}

func registerHTTPServer(lc fx.Lifecycle, cfg config.Config, server *http.Server, logger loglib.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			errCh := make(chan error, 1)
			go func() {
				logger.Info("http server starting", "addr", server.Addr)
				errCh <- server.ListenAndServe()
			}()

			select {
			case err := <-errCh:
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.HTTP.ShutdownTimeout)
			defer cancel()

			logger.Info("http server stopping")
			return server.Shutdown(shutdownCtx)
		},
	})
}
