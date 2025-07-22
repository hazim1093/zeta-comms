package config

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	Logging struct {
		Format string
		Level  string
	}
	Networks map[string]struct {
		ApiUrl       url.URL       `mapstructure:"api_url"`
		PollInterval time.Duration `mapstructure:"poll_interval"`
	} `mapstructure:"networks"`
}

func InitConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		stringToURLHookFunc(),
	)

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, fmt.Errorf("error un-marshalling config: %w", err)
	}
	return &cfg, nil
}

// func to decode string url to URL in config
func stringToURLHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(url.URL{}) {
			return data, nil
		}

		s, ok := data.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string for url.URL")
		}

		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		return *u, nil
	}
}
