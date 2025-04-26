package config

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type LoggerConfig struct {
	zap.Config `yaml:",inline"`
}

func (lc *LoggerConfig) Build() (*zap.Logger, error) {
	return lc.Config.Build()
}

type PoolConfig struct {
	MaxConns          int `yaml:"maxConns"`
	MinConns          int `yaml:"minConns"`
	MaxConnLifetime   int `yaml:"maxConnLifetime"`
	MaxConnIdleTime   int `yaml:"maxConnIdleTime"`
	HealthCheckPeriod int `yaml:"healthCheckPeriod"`
}

type DatabaseConfig struct {
	Host              string     `yaml:"host"`
	Port              int        `yaml:"port"`
	User              string     `yaml:"user"`
	Password          string     `yaml:"password"`
	Dbname            string     `yaml:"dbname"`
	Sslmode           string     `yaml:"sslmode"`
	Schema            string     `yaml:"schema"`
	ConnectRetries    int        `yaml:"connectRetries"`
	ConnectRetryDelay int        `yaml:"connectRetryDelay"`
	Pool              PoolConfig `yaml:"pool"`
}

func (db DatabaseConfig) ConnString() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		db.User, db.Password, db.Host, db.Port, db.Dbname, db.Sslmode, db.Schema,
	)
}

type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	SecretKey string `yaml:"secret_key"`
}

type JwtConfig struct {
	SecretKey  string `yaml:"secret_key"`
	Expiration int    `yaml:"expiration"`
	RefreshExp int    `yaml:"refreshExp"`
}

type Config struct {
	DB             DatabaseConfig       `yaml:"database"`
	Server         ServerConfig         `yaml:"server"`
	Jwt            JwtConfig            `yaml:"jwt"`
	Logger         LoggerConfig         `yaml:"logger"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("could not decode config file: %v", err)
	}
	return &config, nil
}
