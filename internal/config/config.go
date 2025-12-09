package config

import (
	"strings"

	"stockpilot/pkg/flagparser"
	"stockpilot/pkg/gonerve/db"
	"stockpilot/pkg/gonerve/logging"
	"stockpilot/pkg/gonerve/postgresql"
	"stockpilot/pkg/gonerve/sentry"
	"stockpilot/pkg/gonerve/tracing"
)

type Config struct {
	ListenAddr string        `json:"listen_addr" yaml:"listen_addr" flag:"listen-addr" default:":8080" usage:"http listen address"`
	PG         PGConfig      `json:"pg" yaml:"pg" flag:"pg" default:"" usage:"postgres settings"`
	Log        LogConfig     `json:"log" yaml:"log" flag:"log" default:"" usage:"logging settings"`
	Sentry     SentryConfig  `json:"sentry" yaml:"sentry" flag:"sentry" default:"" usage:"sentry settings"`
	Tracing    TracingConfig `json:"tracing" yaml:"tracing" flag:"tracing" default:"" usage:"tracing settings"`
}

type PGConfig struct {
	Endpoint            string `json:"endpoint" yaml:"endpoint" flag:"pg-endpoint" default:"localhost:5432" usage:"postgres host:port"`
	Database            string `json:"database" yaml:"database" flag:"pg-database" default:"stockpilot" usage:"postgres database"`
	Username            string `json:"username" yaml:"username" flag:"pg-username" default:"postgres" usage:"postgres user"`
	Password            string `json:"password" yaml:"password" flag:"pg-password" default:"postgres" usage:"postgres password"`
	SSLMode             string `json:"sslmode" yaml:"sslmode" flag:"pg-sslmode" default:"disable" usage:"postgres sslmode"`
	ListenNotifications bool   `json:"listen_notifications" yaml:"listen_notifications" flag:"pg-listen-notifications" default:"false" usage:"postgres listen notifications"`
}

type LogConfig struct {
	Level             string `json:"level" yaml:"level" flag:"log-level" default:"debug" usage:"log level"`
	DisableCaller     bool   `json:"disable_caller" yaml:"disable_caller" flag:"log-disable-caller" default:"false" usage:"disable caller info"`
	DisableStacktrace bool   `json:"disable_stacktrace" yaml:"disable_stacktrace" flag:"log-disable-stacktrace" default:"false" usage:"disable stacktrace"`
	Output            string `json:"output" yaml:"output" flag:"log-output" default:"stdout" usage:"comma separated log outputs"`
	Encoding          string `json:"encoding" yaml:"encoding" flag:"log-encoding" default:"json" usage:"log encoding"`
	UltraHuman        bool   `json:"ultra_human" yaml:"ultra_human" flag:"log-ultra-human" default:"false" usage:"human friendly logs"`
	LogHTTPRequests   bool   `json:"log_http_requests" yaml:"log_http_requests" flag:"log-http-requests" default:"true" usage:"log http requests"`
}

type SentryConfig struct {
	DSN           string  `json:"dsn" yaml:"dsn" flag:"sentry-dsn" default:"" usage:"sentry dsn"`
	ErrSampleRate float64 `json:"err_sample_rate" yaml:"err_sample_rate" flag:"sentry-sample-rate" default:"1" usage:"sentry error sample rate"`
}

type TracingConfig struct {
	Name        string  `json:"name" yaml:"name" flag:"trace-name" default:"stockpilot" usage:"trace service name"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint" flag:"trace-endpoint" default:"" usage:"otlp collector endpoint"`
	SampleRatio float64 `json:"sample_ratio" yaml:"sample_ratio" flag:"trace-sample" default:"1" usage:"trace sample ratio"`
	Insecure    bool    `json:"insecure" yaml:"insecure" flag:"trace-insecure" default:"true" usage:"otlp insecure transport"`
}

func (c *Config) Load() error {
	return flagparser.ParseFlags(c)
}

func (c PGConfig) ToDBConfig() postgresql.Config {
	options := map[string]any{}
	if c.SSLMode != "" {
		options["sslmode"] = c.SSLMode
	}
	return postgresql.Config{
		Config: db.Config{
			Scheme:   "postgres",
			Driver:   "pgx",
			Endpoint: c.Endpoint,
			Database: c.Database,
			Username: c.Username,
			Password: c.Password,
			Options:  options,
		},
		ListenNotifications: c.ListenNotifications,
	}
}

func (c LogConfig) ToLoggingConfig() logging.Config {
	outputs := []string{}
	for _, v := range strings.Split(c.Output, ",") {
		s := strings.TrimSpace(v)
		if s != "" {
			outputs = append(outputs, s)
		}
	}
	if len(outputs) == 0 {
		outputs = []string{"stdout"}
	}
	return logging.Config{
		Level:             c.Level,
		DisableCaller:     c.DisableCaller,
		DisableStacktrace: c.DisableStacktrace,
		OutputPaths:       outputs,
		Encoding:          c.Encoding,
		UltraHuman:        c.UltraHuman,
		LogHttpRequests:   c.LogHTTPRequests,
	}
}

func (c SentryConfig) ToSentryConfig() *sentry.Config {
	if c.DSN == "" {
		return nil
	}
	return &sentry.Config{
		DSN:           c.DSN,
		ErrSampleRate: c.ErrSampleRate,
	}
}

func (c TracingConfig) ToTracingConfig() *tracing.Config {
	if c.Endpoint == "" {
		return nil
	}
	return &tracing.Config{
		Name:        c.Name,
		SampleRatio: c.SampleRatio,
		Endpoint:    c.Endpoint,
		Insecure:    c.Insecure,
	}
}
