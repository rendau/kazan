package cmd

import (
	"github.com/rendau/dop/dopTools"
	"github.com/spf13/viper"
)

var conf = struct {
	Debug        bool   `mapstructure:"DEBUG"`
	LogLevel     string `mapstructure:"LOG_LEVEL"`
	HttpListen   string `mapstructure:"HTTP_LISTEN"`
	HttpCors     bool   `mapstructure:"HTTP_CORS"`
	SwagHost     string `mapstructure:"SWAG_HOST"`
	SwagBasePath string `mapstructure:"SWAG_BASE_PATH"`
	SwagSchema   string `mapstructure:"SWAG_SCHEMA"`
	ImgMaxWidth  int    `mapstructure:"IMG_MAX_WIDTH"`
	ImgMaxHeight int    `mapstructure:"IMG_MAX_HEIGHT"`
}{}

func confLoad() {
	dopTools.SetViperDefaultsFromObj(conf)

	viper.SetDefault("DEBUG", "false")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("HTTP_LISTEN", ":80")
	viper.SetDefault("SWAG_HOST", "example.com")
	viper.SetDefault("SWAG_BASE_PATH", "/")
	viper.SetDefault("SWAG_SCHEMA", "https")

	viper.SetConfigFile("conf.yml")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	_ = viper.Unmarshal(&conf)
}
