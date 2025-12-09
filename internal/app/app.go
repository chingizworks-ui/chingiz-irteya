package app

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"stockpilot/internal/config"
	"stockpilot/internal/handler"
	"stockpilot/internal/repository/postgres"
	"stockpilot/internal/service"
	"stockpilot/pkg/gonerve/errors"
	"stockpilot/pkg/gonerve/logging"
	"stockpilot/pkg/gonerve/tracing"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	cfgPath := flag.String("config", "config.yaml", "path to config file (yaml or json)")
	flag.Parse()

	cfg := config.Config{}
	if err := loadConfig(*cfgPath, &cfg); err != nil {
		return err
	}

	logCfg := cfg.Log.ToLoggingConfig()
	if err := logging.Init("stockpilot", &logCfg); err != nil {
		return err
	}
	defer logging.Shutdown()

	traceCfg := cfg.Tracing.ToTracingConfig()
	if err := tracing.Init(traceCfg); err != nil {
		return err
	}
	tracing.WithLoggerErrorHandler(logging.GlobalLogger())
	defer tracing.Shutdown(context.Background())

	sentryCfg := cfg.Sentry.ToSentryConfig()
	if sentryCfg != nil {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryCfg.DSN,
			SampleRate:       sentryCfg.ErrSampleRate,
			AttachStacktrace: true,
		}); err != nil {
			return errors.Wrap(err, "sentry init")
		}
		defer sentry.Flush(2 * time.Second)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	repo, err := postgres.New(ctx, cfg.PG.ToDBConfig())
	if err != nil {
		return err
	}
	defer repo.Close()

	userSvc := service.NewUserService(repo)
	productSvc := service.NewProductService(repo)
	orderSvc := service.NewOrderService(repo, repo, repo, repo)

	server, err := handler.NewServer(cfg.ListenAddr, userSvc, productSvc, orderSvc, logCfg.LogHttpRequests, sentryCfg != nil)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.Start(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		logging.Fatal(context.Background(), "server failed", zap.Error(err))
	}

	return nil
}

func loadConfig(path string, cfg *config.Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "read config")
	}
	switch ext := filepath.Ext(path); ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return errors.Wrap(err, "unmarshal yaml")
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return errors.Wrap(err, "unmarshal json")
		}
	default:
		return errors.New("unsupported config format")
	}
	return nil
}
