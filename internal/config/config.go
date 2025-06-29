package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Host                     string `mapstructure:"host"`
	Port                     int    `mapstructure:"port"`
	Environment              string `mapstructure:"environment"`
	LogLevel                 string `mapstructure:"log_level"`
	DatabasePath             string `mapstructure:"database_path"`
	GitLabURL                string `mapstructure:"gitlab_url"`
	RedirectServerURL        string `mapstructure:"redirect_server_url"`
	DefaultWebhookURL        string `mapstructure:"default_webhook_url"`
	GitLabPersonalAccessToken string `mapstructure:"gitlab_personal_access_token"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/gitlab-merge-alert")

	// 设置默认值
	viper.SetDefault("host", "0.0.0.0")
	viper.SetDefault("port", 1688)
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("database_path", "./data/gitlab-merge-alert.db")
	viper.SetDefault("gitlab_url", "https://gitlab.woqutech.com")
	viper.SetDefault("redirect_server_url", "http://localhost:1688")

	// 环境变量绑定
	viper.SetEnvPrefix("GMA")
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// 配置文件不存在，使用默认值
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}