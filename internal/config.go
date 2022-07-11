package internal

import (
	"github.com/spf13/viper"
)

type Config struct {
	RABBIT_MQ_PORT         int    `env:"RABBIT_MQ_PORT"`
	QUEUE_NAME             string `env:"QUEUE_NAME"`
	USERNAME               string `env:"USERNAME"`
	PASSWORD               string `env:"PASSWORD"`
	CLIENT_INPUT_FILENAME  string `env:"CLIENT_INPUT_FILENAME"`
	SERVER_OUTPUT_FILENAME string `env:"SERVER_OUTPUT_FILENAME"`
}

func GetConfig() (*Config, error) {

	viper.SetDefault("RABBIT_MQ_PORT", 5672)
	viper.SetDefault("QUEUE_NAME", "bloxroute")
	viper.SetDefault("USERNAME", "guest")
	viper.SetDefault("PASSWORD", "guest")
	viper.SetDefault("CLIENT_INPUT_FILENAME", "input.json")
	viper.SetDefault("SERVER_OUTPUT_FILENAME", "output.json")
	// read environment variables, will override defaults
	viper.AutomaticEnv()

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
