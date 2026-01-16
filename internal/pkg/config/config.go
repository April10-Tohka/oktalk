package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

func InitConfig() *Config {
	v := viper.New()
	v.SetConfigName("config")    // 文件名
	v.SetConfigType("yaml")      // 扩展名
	v.AddConfigPath("./configs") // 路径

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	if err := v.Unmarshal(&GlobalConfig); err != nil {
		panic(fmt.Errorf("unmarshal config failed: %w", err))
	}

	fmt.Println("✅ 配置中心初始化成功")
	fmt.Printf("%+v\n", GlobalConfig)
	return GlobalConfig
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Aliyun   AliyunConfig   `mapstructure:"aliyun"`
	Xfyun    XfyunConfig    `mapstructure:"xfyun"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type ServerConfig struct {
	Port       int    `mapstructure:"port"`
	Mode       string `mapstructure:"mode"`
	ServerName string `mapstructure:"server_name"`
}

type AliyunConfig struct {
	DASHSCOPE_API_KEY string          `mapstructure:"DASHSCOPE_API_KEY"`
	LLM               AliyunLLMConfig `mapstructure:"LLM"`
	ASR               AliyunASRConfig `mapstructure:"ASR"`
	TTS               AliyunTTSConfig `mapstructure:"TTS"`
}
type AliyunLLMConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}
type AliyunASRConfig struct {
	WsURL string `mapstructure:"ws_url"`
	Model string `mapstructure:"model"`
}
type AliyunTTSConfig struct {
	WsURL string `mapstructure:"ws_url"`
	Model string `mapstructure:"model"`
}
type XfyunConfig struct {
	AppId     string `mapstructure:"app_id"`
	ApiSecret string `mapstructure:"api_secret"`
	ApiKey    string `mapstructure:"api_key"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
