package utils

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver  string `mapstructure:"DB_DRIVER"`
	DBSource  string `mapstructure:"DB_SOURCE"`
	ServerURL string `mapstructure:"SERVER_URL"`
	API       string `mapstructure:"API_VERSION"`
	SECRET    string `mapstructure:"SECRET"`
	// Tiempo de vigencia del token de acceso. Acepta el formato de Go: 30m, 1h, 8h...
	TokenDuration time.Duration `mapstructure:"TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.SetDefault("TOKEN_DURATION", "1h")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}
	if config.TokenDuration <= 0 {
		err = fmt.Errorf("TOKEN_DURATION debe ser mayor que cero, se recibió %q", viper.GetString("TOKEN_DURATION"))
	}
	return
}
