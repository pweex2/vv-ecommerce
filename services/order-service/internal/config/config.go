// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"Host"`
	Port     string `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	DBName   string `mapstructure:"DBName"`
}

// RedisConfig Redis 配置 (占位符)
type RedisConfig struct {
	Addr     string `mapstructure:"Addr"`
	Password string `mapstructure:"Password"`
	DB       int    `mapstructure:"DB"`
}

// MQConfig 消息队列配置 (占位符)
type MQConfig struct {
	Host string `mapstructure:"Host"`
	Port string `mapstructure:"Port"`
}

// Config 应用程序配置
type Config struct {
	ServerPort          int    `mapstructure:"ServerPort"`
	InventoryServiceURL string `mapstructure:"InventoryServiceURL"`

	Database DatabaseConfig `mapstructure:"Database"`
	Redis    RedisConfig    `mapstructure:"Redis"` // 占位符
	MQ       MQConfig       `mapstructure:"MQ"`    // 占位符
}

func LoadConfig() (*Config, error) {
	// 获取当前环境，默认为 development
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// 设置配置文件名和路径
	viper.SetConfigName("config." + env) // 例如: config.development, config.production
	viper.SetConfigType("yaml")          // 配置文件类型
	viper.AddConfigPath("./configs")     // 配置文件搜索路径

	// 从配置文件读取配置
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，可以忽略或记录警告
			fmt.Printf("Warning: Config file for environment '%s' not found, using environment variables or defaults.\n", env)
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 设置默认值 (如果配置文件和环境变量都没有设置)
	viper.SetDefault("ServerPort", 8081)
	viper.SetDefault("Database.Host", "localhost")
	viper.SetDefault("Database.Port", "3306")
	viper.SetDefault("Database.User", "root")
	viper.SetDefault("Database.Password", "root")
	viper.SetDefault("Database.DBName", "order_db")

	// 从环境变量读取配置，环境变量会覆盖配置文件中的值
	viper.AutomaticEnv()

	cfg := &Config{}
	// 将读取到的配置反序列化到 Config 结构体中
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	// 验证配置
	if cfg.ServerPort == 0 {
		return nil, fmt.Errorf("ServerPort cannot be 0")
	}
	if cfg.Database.Host == "" || cfg.Database.Port == "" || cfg.Database.User == "" || cfg.Database.Password == "" || cfg.Database.DBName == "" {
		return nil, fmt.Errorf("Database configuration (Host, Port, User, Password, DBName) cannot be empty")
	}

	return cfg, nil
}

// GetServerPortString 返回 ServerPort 的字符串形式，方便 ListenAndServe 使用
func (c *Config) GetServerPortString() string {
	return ":" + strconv.Itoa(c.ServerPort)
}
