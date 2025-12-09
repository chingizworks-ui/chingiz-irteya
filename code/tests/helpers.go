package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"stockpilot/internal/config"
)

var (
	maxWaiting = time.Second * 10
	maxAttempt = 4
)

func StoreCfgToFile(t *testing.T, cfg config.Config, filePath string) {
	t.Helper()

	extension := strings.ToLower(filepath.Ext(filePath))
	var data []byte
	var err error
	switch extension {
	case ".json":
		data, err = json.Marshal(&cfg)
	case ".yaml", ".yml":
		data, err = yaml.Marshal(&cfg)
	default:
		t.Fatalf("unsupported config extension: %s", extension)
	}
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filePath, data, 0o644))
	t.Cleanup(func() {
		_ = os.Remove(filePath)
	})
}

func EnsureWithRetry(ensure func() error) error {
	timer := time.NewTimer(maxWaiting)
	timeIsOut := false
	go func() {
		<-timer.C
		timeIsOut = true
	}()
	var err error
	for attempt := 0; attempt < maxAttempt; attempt++ {
		if timeIsOut {
			break
		}
		err = ensure()
		if err == nil {
			return nil
		}
		time.Sleep(maxWaiting / time.Duration(maxAttempt))
	}
	return fmt.Errorf("after %d attempts during %s: %v", maxAttempt, maxWaiting, err)
}

func WaitHTTPOKAtAddr(addr string, duration time.Duration) error {
	target := normalizeBaseURL(addr) + "/api/v1/products/unknown-ready-check"
	interval := duration / 20 //nolint:gomnd
	okChan := make(chan struct{})
	timer := time.NewTimer(duration)
	var lastErr error
	lastStatus := 0

	go func() {
		for i := 0; i < 20; i++ {
			r, err := http.Get(target) //nolint:noctx
			if err == nil && r.StatusCode < http.StatusInternalServerError {
				_ = r.Body.Close()
				close(okChan)
				return
			}
			lastErr = err
			if r != nil {
				lastStatus = r.StatusCode
			}
			time.Sleep(interval)
		}
	}()

	select {
	case <-okChan:
		return nil
	case <-timer.C:
		return fmt.Errorf("address %s doesn't respond at %s (last status %d, last error %v)", addr, duration, lastStatus, lastErr)
	}
}

func LoadConfigFromFile(t *testing.T, path string) *config.Config {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	cfg := config.Config{}
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".json":
		err = json.Unmarshal(data, &cfg)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &cfg)
	default:
		t.Fatalf("unsupported config extension: %s", ext)
	}
	require.NoError(t, err)
	return &cfg
}
