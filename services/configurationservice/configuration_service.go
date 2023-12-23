package configurationservice

import (
	"os"
	"strconv"

	"github.com/MorrisMorrison/gchat/logger"
)

func GetBaseUrl() string {
	baseUrl := os.Getenv("GCHAT_BASE_URL")
	if baseUrl == "" {
		logger.Log.Info("Env variable GCHAT_BASE_URL is not set. Use default host localhost.")
		return "localhost"
	}

	return baseUrl
}

func GetPort() string {
	port := os.Getenv("GCHAT_PORT")
	if port == "" {
		logger.Log.Info("Env variable GCHAT_PORT is not set. Use default port 8080.")
		return "8080"
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		logger.Log.Infof("Env variable GCHAT_PORT %s is not a valid integer. Use default port 8080.", port)
		return "8080"
	}

	if portNumber > 65536 {
		logger.Log.Infof("Provided port number %d is invalid, because it is larger than 65536. Use default port 8080.", portNumber)
		return "8080"
	}

	return port
}
