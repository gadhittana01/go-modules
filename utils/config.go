package utils

import (
	"time"

	"github.com/spf13/viper"
)

type BaseConfig struct {
	ServerPort          int           `mapstructure:"SERVER_PORT"`
	DBConnString        string        `mapstructure:"DB_CONN_STRING"`
	DBName              string        `mapstructure:"DB_NAME"`
	MigrationURL        string        `mapstructure:"MIGRATION_URL"`
	JWTKey              string        `mapstructure:"JWT_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RedisHost           string        `mapstructure:"REDIS_HOST"`
	RedisUsername       string        `mapstructure:"REDIS_USERNAME"`
	RedisPassword       string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB             int           `mapstructure:"REDIS_DB"`
	CacheDuration       time.Duration `mapstructure:"CACHE_DURATION"`
}

func LoadBaseConfig(path string, configName string, config *BaseConfig) {
	viper.AddConfigPath(path)
	viper.SetConfigName(configName)
	viper.SetTypeByDefaultValue(true)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
}

func CheckAndSetConfig(path string, configName string) *BaseConfig {
	config := &BaseConfig{}
	LoadBaseConfig(path, configName, config)
	return config
}
