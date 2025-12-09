package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"stockpilot/internal/config"
	"stockpilot/internal/handler"
	"stockpilot/internal/service"
	"stockpilot/pkg/gonerve/logging"
)

type Suite struct {
	Config        config.Config
	ApiClient     *Client
	Server        *handler.Server
	Repo          *MemoryRepository
	GetServerLogs func() ([]string, error)
}

func CreateUnitTestingSuite(t *testing.T, cfg *config.Config) *Suite {
	t.Helper()

	logCfg := cfg.Log.ToLoggingConfig()
	require.NoError(t, logging.Init("stockpilot-tests", &logCfg))

	repo := NewMemoryRepository()

	userSvc := service.NewUserService(repo)
	productSvc := service.NewProductService(repo)
	orderSvc := service.NewOrderService(repo, repo, repo, repo)

	server, err := handler.NewServer(cfg.ListenAddr, userSvc, productSvc, orderSvc, cfg.Log.LogHTTPRequests, cfg.Sentry.ToSentryConfig() != nil)
	require.NoError(t, err)

	go func() {
		if errStart := server.Start(); errStart != nil && !errors.Is(errStart, http.ErrServerClosed) {
			t.Errorf("server start failed: %v", errStart)
		}
	}()

	require.NoError(t, WaitHTTPOKAtAddr(cfg.ListenAddr, maxWaiting))

	suite := &Suite{
		Config:        *cfg,
		ApiClient:     NewAPIClient(*cfg),
		Server:        server,
		Repo:          repo,
		GetServerLogs: func() ([]string, error) { return []string{}, nil },
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		_ = logging.Shutdown()
	})

	return suite
}
