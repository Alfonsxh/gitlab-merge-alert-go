package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Host                      string `mapstructure:"host"`
	Port                      int    `mapstructure:"port"`
	Environment               string `mapstructure:"environment"`
	LogLevel                  string `mapstructure:"log_level"`
	DatabasePath              string `mapstructure:"database_path"`
	GitLabURL                 string `mapstructure:"gitlab_url" json:"-"` // 敏感字段不输出到日志
	PublicWebhookURL          string `mapstructure:"public_webhook_url"`
	GitLabPersonalAccessToken string `mapstructure:"gitlab_personal_access_token" json:"-"` // 敏感字段不输出到日志
}

// MaskSensitive 返回一个掩码后的配置副本，用于日志输出
func (c *Config) MaskSensitive() Config {
	masked := *c
	if masked.GitLabURL != "" {
		masked.GitLabURL = maskURL(masked.GitLabURL)
	}
	if masked.GitLabPersonalAccessToken != "" {
		masked.GitLabPersonalAccessToken = "****"
	}
	return masked
}

func maskURL(url string) string {
	if url == "" {
		return ""
	}
	if strings.Contains(url, "://") {
		parts := strings.SplitN(url, "://", 2)
		if len(parts) == 2 && len(parts[1]) > 4 {
			hostPart := parts[1]
			if len(hostPart) > 8 {
				return parts[0] + "://" + hostPart[:3] + "****" + hostPart[len(hostPart)-3:]
			}
		}
	}
	return "****"
}

func Load() (*Config, error) {
	// 配置文件查找优先级：config.local.yaml > config.yaml > 环境变量
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/gitlab-merge-alert")

	// 设置非敏感信息的默认值
	viper.SetDefault("host", "0.0.0.0")
	viper.SetDefault("port", 1688)
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("database_path", "./data/gitlab-merge-alert.db")

	// 环境变量绑定（优先级最高）
	viper.SetEnvPrefix("GMA")
	viper.AutomaticEnv()

	// 尝试读取本地配置文件（优先级高）
	viper.SetConfigName("config.local")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 本地配置不存在，尝试读取默认配置文件
			viper.SetConfigName("config")
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return nil, fmt.Errorf("读取配置文件失败: %w", err)
				}
				// 所有配置文件都不存在，依赖环境变量和默认值
			}
		} else {
			return nil, fmt.Errorf("读取本地配置文件失败: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必要的敏感配置
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig 验证配置的完整性
func validateConfig(config *Config) error {
	var missingFields []string

	if config.GitLabURL == "" || config.GitLabURL == "https://your-gitlab-server.com" {
		missingFields = append(missingFields, "gitlab_url (GitLab服务器地址)")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("缺少必要的配置项: %s\n\n请在以下位置之一配置这些敏感信息:\n1. 创建 config.local.yaml 文件\n2. 设置环境变量 (GMA_GITLAB_URL)\n3. 参考 config.example.yaml 了解配置格式", strings.Join(missingFields, ", "))
	}

	return nil
}
