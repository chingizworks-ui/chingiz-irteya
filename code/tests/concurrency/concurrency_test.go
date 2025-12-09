package concurrency

import (
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"stockpilot/code/tests"
	"stockpilot/code/tests/configs/ports"
	"stockpilot/internal/config"
)

var TestSuite *tests.Suite

func Test_SpecConcurrency(t *testing.T) {
	cfg := baseCfg(t)
	cfg.ListenAddr = fmt.Sprintf(":%d", ports.ConcurrencyHTTPPort)

	TestSuite = tests.CreateUnitTestingSuite(t, cfg)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stockpilot Concurrency Suite")
}

func baseCfg(t *testing.T) *config.Config {
	cfgPath := filepath.Clean("../configs/base-config.yaml")
	return tests.LoadConfigFromFile(t, cfgPath)
}
