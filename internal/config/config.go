package config

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ConfigFile       string
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
)

func InitConfig() {
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".innominatusrchestrator")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "idp_orchestrator")
	viper.SetDefault("database.sslmode", "disable")

	if DatabaseHost == "" {
		DatabaseHost = viper.GetString("database.host")
	}
	if DatabasePort == 0 {
		DatabasePort = viper.GetInt("database.port")
	}
	if DatabaseUser == "" {
		DatabaseUser = viper.GetString("database.user")
	}
	if DatabasePassword == "" {
		DatabasePassword = viper.GetString("database.password")
	}
	if DatabaseName == "" {
		DatabaseName = viper.GetString("database.name")
	}

	if DatabasePassword == "" {
		if envPassword := os.Getenv("POSTGRES_PASSWORD"); envPassword != "" {
			DatabasePassword = envPassword
		}
	}
}
