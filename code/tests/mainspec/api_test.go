package mainspec

import (
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	"stockpilot/code/tests"
	"stockpilot/code/tests/configs/ports"
	"stockpilot/internal/config"
)

var TestSuite *tests.Suite

func Test_SpecStockpilot(t *testing.T) {
	cfg := baseCfg(t)
	cfg.ListenAddr = fmt.Sprintf(":%d", ports.BaseHTTPPort)

	TestSuite = tests.CreateUnitTestingSuite(t, cfg)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stockpilot Base Suite")
}

func baseCfg(t *testing.T) *config.Config {
	cfgPath := filepath.Clean("../configs/base-config.yaml")
	cfg := tests.LoadConfigFromFile(t, cfgPath)
	require.NotEmpty(t, cfg.ListenAddr)
	return cfg
}
