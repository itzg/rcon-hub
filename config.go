package rcon_hub

import (
	"errors"
	"github.com/spf13/viper"
	"strings"
)

const (
	ConfigHistorySize = "history-size"
	ConfigBind        = "bind"
	ConfigHostKeyFile = "host-key-file"
	ConfigUser        = "user"
	ConfigPassword    = "password"
	ConfigConnection  = "connection"

	ConnectionFormat = "name=password@host:port"
)

type User struct {
	Password string
}

type Connection struct {
	Address  string
	Password string
}

type Config struct {
	HistorySize int    `mapstructure:"history-size"`
	HostKeyFile string `mapstructure:"host-key-file"`
	Users       map[string]User
	Connections map[string]Connection
}

func (c *Config) AddExtraConnection(s string) error {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return errors.New("invalid name part of connection")
	}

	subParts := strings.SplitN(parts[1], "@", 2)
	if len(subParts) != 2 {
		return errors.New("invalid connection definition")
	}

	c.Connections[parts[0]] = Connection{
		Address:  subParts[1],
		Password: subParts[0],
	}

	return nil
}

func LoadConfig() (*Config, error) {
	config := new(Config)

	err := viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	extraUser := viper.GetString(ConfigUser)
	extraPassword := viper.GetString(ConfigPassword)
	if extraPassword != "" && extraUser != "" {
		if config.Users == nil {
			config.Users = make(map[string]User)
		}
		config.Users[extraUser] = User{Password: extraPassword}
	}
	if len(config.Users) == 0 {
		return nil, errors.New("no users declared")
	}

	if config.Connections == nil {
		config.Connections = make(map[string]Connection)
	}
	for _, extraConn := range viper.GetStringSlice(ConfigConnection) {
		err := config.AddExtraConnection(extraConn)
		if err != nil {
			return nil, err
		}
	}
	if len(config.Connections) == 0 {
		return nil, errors.New("no connections declared")
	}

	return config, nil
}
