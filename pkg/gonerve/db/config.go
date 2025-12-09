package db

import (
	"fmt"
	"strings"

	"stockpilot/pkg/gonerve/errors"
)

type Config struct {
	Scheme   string
	Driver   string
	Endpoint string
	Database string
	Username string
	Password string
	Options  map[string]any
}

func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint cannot be empty")
	}
	if c.Database == "" {
		return errors.New("database cannot be empty")
	}
	if c.Username == "" {
		return errors.New("username cannot be empty")
	}
	if c.Scheme == "" {
		c.Scheme = "postgres"
	}
	if c.Driver == "" {
		c.Driver = "pgx"
	}
	return nil
}

func (c Config) ToConnString() string {
	var b strings.Builder
	if c.Scheme == "" {
		c.Scheme = "postgres"
	}
	b.WriteString(c.Scheme)
	b.WriteString("://")
	b.WriteString(c.Username)
	if c.Password != "" {
		b.WriteString(":")
		b.WriteString(c.Password)
	}
	b.WriteString("@")
	b.WriteString(c.Endpoint)
	b.WriteString("/")
	b.WriteString(c.Database)
	if len(c.Options) > 0 {
		b.WriteString("?")
		parts := make([]string, 0, len(c.Options))
		for k, v := range c.Options {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		b.WriteString(strings.Join(parts, "&"))
	}
	return b.String()
}
