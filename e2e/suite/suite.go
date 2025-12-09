package suite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"stockpilot/code/tests"
	"stockpilot/internal/config"
)

const defaultWaitHTTPOKTime = 10 * time.Second

type PostgresInstance struct {
	container  tc.Container
	ConnString string
	Endpoint   string
}

type E2ESuite struct {
	name       string
	configFile string
	TestSuite  *tests.Suite
	cmd        *exec.Cmd
	prc        *os.Process
	outb, errb bytes.Buffer
	stopped    bool
	pg         *PostgresInstance
}

func (e *E2ESuite) StopWithLogs() {
	if e.stopped {
		return
	}
	e.stopped = true
	e.GracefulStop()
	fmt.Printf("\n[%s] e2e tests server logs & output:\n ========= \n%s\n%s", e.name, e.errb.String(), e.outb.String())
	if e.pg != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = e.pg.container.Terminate(ctx)
	}
}

func (e *E2ESuite) GracefulStop() {
	if e.prc == nil {
		return
	}
	_ = e.prc.Signal(syscall.SIGTERM)
	_, _ = e.prc.Wait()
}

func (e *E2ESuite) Run(t *testing.T) {
	e.stopped = false
	var serverError error
	defer func() {
		if serverError != nil {
			fmt.Printf("\nSERVER START FAILURE:\n%s\n%s", e.errb.String(), e.outb.String())
			t.Fatal(serverError)
		}
	}()

	go func() {
		fmt.Println("start server:", e.name)
		binaryPath := filepath.Join("..", "..", "binaries", "stockpilot-service-native")
		e.cmd = exec.Command(binaryPath, "-config", e.configFile)
		e.cmd.Stdout = &e.outb
		e.cmd.Stderr = &e.errb
		e.cmd.Stdin = os.Stdin
		serverError = e.cmd.Start()
		e.prc = e.cmd.Process
		if serverError != nil {
			return
		}
		serverError = e.cmd.Wait()
		if serverError != nil {
			fmt.Printf("starting error: %v", serverError)
		}
	}()

	waitingError := tests.WaitHTTPOKAtAddr(e.TestSuite.Config.ListenAddr, defaultWaitHTTPOKTime)

	if waitingError != nil {
		serverError = waitingError
		return
	}

	e.prc = e.cmd.Process
}

type Params struct {
	Cfg            config.Config
	TestName       string
	ConfigFilePath string
}

func RunNewE2ESuite(t *testing.T, params Params) *E2ESuite {
	t.Helper()

	pg := startPostgresContainer(t)
	params.Cfg.PG.Endpoint = pg.Endpoint
	params.Cfg.PG.Database = "stockpilot"
	params.Cfg.PG.Username = "postgres"
	params.Cfg.PG.Password = "postgres"
	params.Cfg.PG.SSLMode = "disable"

	migrationsPath := filepath.Clean("../../migrations/0001_init.sql")
	require.NoError(t, tests.ApplyMigrations(context.Background(), pg.ConnString, migrationsPath))

	tests.StoreCfgToFile(t, params.Cfg, params.ConfigFilePath)

	var e2eSuite *E2ESuite

	suite := tests.Suite{
		Config:    params.Cfg,
		ApiClient: tests.NewAPIClient(params.Cfg),
		GetServerLogs: func() ([]string, error) {
			rawLogs := e2eSuite.outb.String() + e2eSuite.errb.String()
			return strings.Split(rawLogs, "\n"), nil
		},
	}
	e2eSuite = &E2ESuite{
		name:       params.TestName,
		configFile: params.ConfigFilePath,
		TestSuite:  &suite,
		pg:         pg,
	}

	e2eSuite.Run(t)

	return e2eSuite
}

func startPostgresContainer(t *testing.T) *PostgresInstance {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req := tc.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "stockpilot",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(40 * time.Second),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mapped, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	endpoint := fmt.Sprintf("%s:%s", host, mapped.Port())
	connString := fmt.Sprintf("postgres://postgres:postgres@%s/stockpilot?sslmode=disable", endpoint)

	return &PostgresInstance{
		container:  container,
		ConnString: connString,
		Endpoint:   endpoint,
	}
}
