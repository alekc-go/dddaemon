package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.alekc.dev/dddaemon"
)

func main() {
	// config block
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/dddaemon/")
	viper.AddConfigPath("$HOME/.dddaemon/")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			//
			panic(err)
		}
	}

	// setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
	dddaemon.Server{}.Run()
}
