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

type Config struct {
	ServerPort int            `mapstructure:"ServerPort"`
	Database   DatabaseConfig `mapstructure:"Database"`
}

func LoadConfig() (*Config, error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	viper.SetConfigName("config." + env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs") // ??? cmd ???????

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Warning: Config file for environment '%s' not found.\n", env)
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	viper.SetDefault("ServerPort", 8083)
	viper.SetDefault("Database.Host", "localhost")
	viper.SetDefault("Database.Port", "3306")
	viper.SetDefault("Database.User", "root")
	viper.SetDefault("Database.Password", "root")
	viper.SetDefault("Database.DBName", "payment_db")

	// Allow environment variables to override config, replacing . with _
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return cfg, nil
}
