package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	NewRelicKey        string `mapstructure:"NEW_RELIC_KEY"`
	NewRelicLicenseKey string `mapstructure:"NEW_RELIC_LICENSE_KEY"`
	AppName            string `mapstructure:"APP_NAME"`
	Port               string `mapstructure:"PORT"`
}

var AppConfig *Config

func init() {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	FailOnError(v.BindEnv("NEW_RELIC_KEY"), "failed to bind NEW_RELIC_KEY")
	FailOnError(v.BindEnv("NEW_RELIC_LICENSE_KEY"), "failed to bind NEW_RELIC_LICENSE_KEY")
	FailOnError(v.BindEnv("APP_NAME"), "failed to bind APP_NAME")
	FailOnError(v.BindEnv("PORT"), "failed to bind PORT")
	err := v.ReadInConfig()
	if err != nil {
		log.Println("Load from environment variable")
	}
	err = v.Unmarshal(&AppConfig)
	if err != nil {
		FailOnError(err, "Failed to read enivronment")
	}
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
