package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Mongo  MongoConfig  `mapstructure:"mongo"`
	App    AppConfig    `mapstructure:"app"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type MongoConfig struct {
	URI    string `mapstructure:"uri"`
	DBName string `mapstructure:"db_name"`
}

type AppConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

func LoadConfig(path string) (config *Config, err error) {
	var cfg Config
	if err := ReadConfig(filepath.Join(path, "config.yml"), &cfg); err == nil {
		return &cfg, nil
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	v.AddConfigPath(filepath.Join(path, "config"))

	v.AutomaticEnv()

	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err = v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReadConfig(path string, out interface{}) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(out)
}
