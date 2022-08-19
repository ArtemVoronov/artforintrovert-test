package utils

import (
	"log"
	"os"
)

func EnvVar(varName string) string {
	val, valExists := os.LookupEnv(varName)
	if !valExists {
		log.Fatalf("Missed enviroment variable: %s. Check the .env file or OS enviroment vars", varName)
	}
	return val
}

func EnvVarDefault(varName string, defaultValue string) string {
	val, valExists := os.LookupEnv(varName)
	if !valExists {
		return defaultValue
	}
	return val
}
