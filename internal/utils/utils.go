package utils

import (
	"log"
	"os"
	"strconv"
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

func EnvVarIntDefault(varName string, defaultValue string) int {
	val := EnvVarDefault(varName, defaultValue)
	result, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Unable to parse enviroment variable: %s. Using default value", varName)
	}
	return result
}
