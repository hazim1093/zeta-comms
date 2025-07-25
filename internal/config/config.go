package config

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Networks map[string]struct {
		ApiUrl       url.URL       `mapstructure:"api_url"`
		PollInterval time.Duration `mapstructure:"poll_interval"`
		Audiences    []string      `mapstructure:"audiences"`
	} `mapstructure:"networks"`

	AudienceConfig map[string]struct {
		Channels map[string][]string `mapstructure:"channels"`
	} `mapstructure:"audience_config"`

	Events struct {
		Proposals struct {
			Filters struct {
				MessageTypes []string `mapstructure:"message_types"`
			} `mapstructure:"filters"`
		} `mapstructure:"proposals"`
	} `mapstructure:"events"`

	Notifiers struct {
		Discord struct {
			BotToken string `mapstructure:"bot_token"`
		} `mapstructure:"discord"`

		Telegram struct {
			BotToken string `mapstructure:"bot_token"`
		} `mapstructure:"telegram"`
	} `mapstructure:"notifiers"`

	Storage struct {
		Filename string `mapstructure:"filename"`
	} `mapstructure:"storage"`

	Logging struct {
		Format string
		Level  string
	}
}

func InitConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set up command line flags
	pflag.String("config", "", "Additional config files to load (comma-separated)")
	pflag.Parse()

	err := v.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("error binding flags: %w", err)
	}

	// Read the base config file first
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading base config file: %w", err)
	}

	// Check if additional config files were specified
	if configFiles := v.GetString("config"); configFiles != "" {
		// Split comma-separated list of config files
		for _, configFile := range strings.Split(configFiles, ",") {
			v.SetConfigFile(configFile)

			if err := v.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("error merging config file %s: %w", configFile, err)
			}
		}
	}

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		stringToURLHookFunc(),
		envVarInterpolationHookFunc(), // Add environment variable interpolation
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

// Function to process environment variables in strings
func envVarInterpolationHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		// Only process strings
		if f.Kind() != reflect.String {
			return data, nil
		}

		str, ok := data.(string)
		if !ok {
			return data, nil
		}

		// Look for ${VAR} pattern and replace with environment variable
		result := os.Expand(str, func(key string) string {
			value, exists := os.LookupEnv(key)
			if !exists {
				// Return the original ${VAR} if environment variable doesn't exist
				return "${" + key + "}"
			}

			return value
		})

		return result, nil
	}
}
