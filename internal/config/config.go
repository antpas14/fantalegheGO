// config/config.go
package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
    BaseUrl string `yaml:"baseUrl"`
    RankingUrl string `yaml:"rankingUrl"`
    CalendarUrl string `yaml:"calendarUrl"`
	FetcherUrl string `yaml:"fetcherUrl"`
}

func LoadConfig() (Config, error) {
	var config Config

	baseConfig, err := loadConfig("config.yaml")
	if err != nil {
		return config, err
	}
	isDocker := os.Getenv("DOCKER_ENV") == "true"

	if isDocker {
		dockerConfig, err := loadConfig("config_docker.yaml")
		if err != nil {
			return config, err
		}

		// Overwrite FetcherUrl config if is in docker
		if dockerConfig.FetcherUrl != "" {
			baseConfig.FetcherUrl = dockerConfig.FetcherUrl
		}

		config = baseConfig
	} else {
		// Use the base configuration
		config = baseConfig
	}

	return config, nil
}

func loadConfig(filename string) (Config, error) {
	var config Config

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}