package db

import (
	"time"

	"github.com/ArtemVoronov/artforintrovert-test/internal/utils"
)

func ConnectTimeout() time.Duration {
	value := utils.EnvVarIntDefault("DATABASE_CONNECT_TIMEOUT_IN_SECONDS", "30")
	return time.Duration(value) * time.Second
}

func QueryTimeout() time.Duration {
	value := utils.EnvVarIntDefault("DATABASE_QUERY_TIMEOUT_IN_SECONDS", "30")
	return time.Duration(value) * time.Second
}

func DBName() string {
	return utils.EnvVarDefault("DATABASE_NAME", "testdb")
}
