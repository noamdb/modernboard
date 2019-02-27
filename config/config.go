package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.SetConfigName("default-conf")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}
	viper.SetConfigName("conf")
	err = viper.MergeInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}
}
