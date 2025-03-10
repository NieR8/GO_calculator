package env

import (
	"os"
	"strconv"
)

// Cодержит конфигурацию приложения, загруженную из переменных окружения среды
type Config struct {
	ComputingPower       int
	TimeAdditionMS       int
	TimeSubtractionMS    int
	TimeMultiplicationMS int
	TimeDivisionMS       int
	OrchestratorAddr     string
}

// Загружает конфигурацию из переменных окружения
func LoadConfig() Config {
	return Config{
		ComputingPower:       getEnvInt("COMPUTING_POWER", 3),
		TimeAdditionMS:       getEnvInt("TIME_ADDITION_MS", 200),
		TimeSubtractionMS:    getEnvInt("TIME_SUBTRACTION_MS", 150),
		TimeMultiplicationMS: getEnvInt("TIME_MULTIPLICATIONS_MS", 100),
		TimeDivisionMS:       getEnvInt("TIME_DIVISIONS_MS", 250),
		OrchestratorAddr:     getEnvString("ORCHESTRATOR_ADDR", ":8080"),
	}
}

// Читает переменную окружения и возвращает ее с дефолтным значением
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Читает переменную окружения и также возвращает ее с дефолтным значением
func getEnvString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
