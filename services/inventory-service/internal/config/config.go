package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"Host"`
	Port     string `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	DBName   string `mapstructure:"DBName"`
}

type RedisConfig struct {
	Host     string `mapstructure:"Host"`
	Port     string `mapstructure:"Port"`
	Password string `mapstructure:"Password"`
	DB       int    `mapstructure:"DB"`
}

type MQConfig struct {
	Host     string `mapstructure:"Host"`
	Port     string `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
}

type Config struct {
	ServerPort int            `mapstructure:"ServerPort"`
	Database   DatabaseConfig `mapstructure:"Database"`
	Redis      RedisConfig    `mapstructure:"Redis"`
	MQ         MQConfig       `mapstructure:"MQ"`
}

func LoadConfig() (*Config, error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	viper.SetConfigName("config." + env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	viper.SetDefault("ServerPort", 8080)
	viper.SetDefault("Database.Host", "localhost")
	viper.SetDefault("Database.Port", "3306")
	viper.SetDefault("Database.User", "root")
	viper.SetDefault("Database.Password", "root")
	viper.SetDefault("Database.DBName", "inventory_db")

	viper.SetDefault("Redis.Host", "localhost")
	viper.SetDefault("Redis.Port", "6379")
	viper.SetDefault("Redis.Password", "")
	viper.SetDefault("Redis.DB", 0)

	viper.SetDefault("MQ.Host", "localhost")
	viper.SetDefault("MQ.Port", "5672")
	viper.SetDefault("MQ.User", "guest")
	viper.SetDefault("MQ.Password", "guest")

	// Allow environment variables to override config, replacing . with _
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	// 验证配置
	if cfg.ServerPort == 0 {
		return nil, fmt.Errorf("ServerPort cannot be 0")
	}
	if cfg.Database.Host == "" || cfg.Database.Port == "" || cfg.Database.User == "" || cfg.Database.DBName == "" {
		return nil, fmt.Errorf("Database configuration (Host, Port, User, DBName) cannot be empty")
	}

	return cfg, nil
}
