package config

import "github.com/spf13/viper"

func InitConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

type ElectrumConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Queue       string `mapstructure:"queue"`
	QueueType   string `mapstructure:"queue_type"`
	RoutingKey  string `mapstructure:"routing_key"`
	SourceChain string `mapstructure:"source_chain"`
	Enabled     *bool  `mapstructure:"enabled"`
	StopHeight  *int64 `mapstructure:"stop_height"`
}

var Config *ElectrumConfig

func LoadEnv() error {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig() error {
	Config = &ElectrumConfig{}
	err := viper.Unmarshal(Config)
	if err != nil {
		return err
	}
	return nil
}
