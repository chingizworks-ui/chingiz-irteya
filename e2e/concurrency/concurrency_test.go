package concurrency

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	"stockpilot/code/tests"
	concurrencySpec "stockpilot/code/tests/concurrency"
	"stockpilot/code/tests/configs/ports"
	"stockpilot/e2e/suite"
	"stockpilot/internal/config"
)

func Test_E2EConcurrencyStockpilot(t *testing.T) {
	cfg := baseCfg(t)

	tmpDir, err := os.MkdirTemp("", "stockpilot-e2e-concurrency-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	cfg.ListenAddr = fmt.Sprintf(":%d", ports.E2EConcurrencyHTTPPort)

	configPath := filepath.Join(tmpDir, "config.yaml")

	e2eSuite := suite.RunNewE2ESuite(t, suite.Params{
		Cfg:            *cfg,
		TestName:       "E2E Stockpilot Concurrency",
		ConfigFilePath: configPath,
	})
	defer e2eSuite.StopWithLogs()

	concurrencySpec.TestSuite = e2eSuite.TestSuite
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Stockpilot Concurrency Suite")
}

func baseCfg(t *testing.T) *config.Config {
	cfgPath := filepath.Clean("../../code/tests/configs/base-config.yaml")
	cfg := tests.LoadConfigFromFile(t, cfgPath)
	require.NotEmpty(t, cfg.ListenAddr)
	return cfg
}
