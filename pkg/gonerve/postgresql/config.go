package postgresql

import "stockpilot/pkg/gonerve/db"

type Config struct {
	db.Config
	ListenNotifications bool `mapstructure:"listen_notifications" json:"listen_notifications" yaml:"listen_notifications"`
}

func (c Config) Validate() error {
	return c.Config.Validate()
}

func (c Config) GetDriver() string { return c.Driver }

func (c Config) ToConnString() string {
	if c.Scheme == "" {
		c.Scheme = "postgres"
	}
	return c.Config.ToConnString()
}
