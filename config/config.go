package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panic("Config not found...")
	}
}
