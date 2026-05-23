package factory

import (
	"fmt"

	"speaknow/internal/config"
	"speaknow/internal/provider"
	"speaknow/internal/provider/aliyun"
	"speaknow/internal/provider/mock"
	"speaknow/internal/provider/tencent"
	"speaknow/internal/provider/xunfei"
)

func BuildRegistry(cfg *config.Config) (*provider.Registry, error) {
	var providers []provider.Provider

	if cfg.Providers.Mock.Enabled {
		providers = append(providers, mock.New(cfg.Providers.Mock.CostPerSecond))
	}
	if cfg.Providers.Aliyun.Enabled {
		providers = append(providers, aliyun.New(
			cfg.Providers.Aliyun.AppKey,
			cfg.Providers.Aliyun.AccessKeyID,
			cfg.Providers.Aliyun.AccessKeySecret,
			cfg.Providers.Aliyun.CostPerSecond,
		))
	}
	if cfg.Providers.Tencent.Enabled {
		providers = append(providers, tencent.New(
			cfg.Providers.Tencent.SecretID,
			cfg.Providers.Tencent.SecretKey,
			cfg.Providers.Tencent.AppID,
			cfg.Providers.Tencent.CostPerSecond,
		))
	}
	if cfg.Providers.Xunfei.Enabled {
		providers = append(providers, xunfei.New(
			cfg.Providers.Xunfei.AppID,
			cfg.Providers.Xunfei.APIKey,
			cfg.Providers.Xunfei.APISecret,
			cfg.Providers.Xunfei.HostURL,
			cfg.Providers.Xunfei.CostPerSecond,
		))
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no ASR provider enabled")
	}
	return provider.NewRegistry(providers...), nil
}
