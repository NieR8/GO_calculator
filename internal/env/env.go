package env

import (
	"os"
	"strconv"
)

// Config содержит конфигурацию приложения, загруженную из переменных окружения
type Config struct {
	ComputingPower       int
	TimeAdditionMS       int
	TimeSubtractionMS    int
	TimeMultiplicationMS int
	TimeDivisionMS       int
	OrchestratorAddr     string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() Config {
	return Config{
		ComputingPower:       getEnvInt("COMPUTING_POWER", 1),
		TimeAdditionMS:       getEnvInt("TIME_ADDITION_MS", 100),
		TimeSubtractionMS:    getEnvInt("TIME_SUBTRACTION_MS", 100),
		TimeMultiplicationMS: getEnvInt("TIME_MULTIPLICATIONS_MS", 100),
		TimeDivisionMS:       getEnvInt("TIME_DIVISIONS_MS", 100),
		OrchestratorAddr:     getEnvString("ORCHESTRATOR_ADDR", ":8080"),
	}
}

// getEnvInt читает переменную окружения и возвращает int, с дефолтным значением
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvString читает переменную окружения и возвращает string, с дефолтным значением
func getEnvString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
