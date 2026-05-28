package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Cache     CacheConfig     `mapstructure:"cache"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	ASR       ASRConfig       `mapstructure:"asr"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Log       LogConfig       `mapstructure:"log"`
}

type ServerConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	Mode              string        `mapstructure:"mode"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
}

type CacheConfig struct {
	TTL time.Duration `mapstructure:"ttl"`
}

type RateLimitConfig struct {
	GlobalQPS   float64 `mapstructure:"global_qps"`
	GlobalBurst int     `mapstructure:"global_burst"`
	UserQPS     float64 `mapstructure:"user_qps"`
	UserBurst   int     `mapstructure:"user_burst"`
}

type ASRConfig struct {
	Primary      string        `mapstructure:"primary"`
	Fallback     []string      `mapstructure:"fallback"`
	Timeout      time.Duration `mapstructure:"timeout"`
	MaxAudioSize int64         `mapstructure:"max_audio_size"`
	MaxDuration  time.Duration `mapstructure:"max_duration"`
}

type ProvidersConfig struct {
	Mock    MockProviderConfig    `mapstructure:"mock"`
	Aliyun  AliyunProviderConfig  `mapstructure:"aliyun"`
	Tencent TencentProviderConfig `mapstructure:"tencent"`
	Xunfei  XunfeiProviderConfig  `mapstructure:"xunfei"`
	Vosk    VoskProviderConfig    `mapstructure:"vosk"`
}

type MockProviderConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	CostPerSecond float64 `mapstructure:"cost_per_second"`
}

type AliyunProviderConfig struct {
	Enabled         bool    `mapstructure:"enabled"`
	AppKey          string  `mapstructure:"app_key"`
	AccessKeyID     string  `mapstructure:"access_key_id"`
	AccessKeySecret string  `mapstructure:"access_key_secret"`
	CostPerSecond   float64 `mapstructure:"cost_per_second"`
}

type TencentProviderConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	SecretID      string  `mapstructure:"secret_id"`
	SecretKey     string  `mapstructure:"secret_key"`
	AppID         string  `mapstructure:"app_id"`
	CostPerSecond float64 `mapstructure:"cost_per_second"`
}

type XunfeiProviderConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	AppID         string  `mapstructure:"app_id"`
	APIKey        string  `mapstructure:"api_key"`
	APISecret     string  `mapstructure:"api_secret"`
	HostURL       string  `mapstructure:"host_url"`
	CostPerSecond float64 `mapstructure:"cost_per_second"`
}

type VoskProviderConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	ModelPath     string  `mapstructure:"model_path"`
	SampleRate    float64 `mapstructure:"sample_rate"`
	CostPerSecond float64 `mapstructure:"cost_per_second"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("SPEAKNOW")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.ReadHeaderTimeout <= 0 {
		c.Server.ReadHeaderTimeout = 10 * time.Second
	}
	// ReadTimeout / WriteTimeout 为 0 表示不限制，便于 WebSocket 长连接。
	if c.Server.IdleTimeout <= 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}
	if c.Server.ShutdownTimeout <= 0 {
		c.Server.ShutdownTimeout = 30 * time.Second
	}
	if c.Cache.TTL <= 0 {
		c.Cache.TTL = 24 * time.Hour
	}
}
