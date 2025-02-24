package main

import (
	"context"
	"github.com/NieR8/myProject/agent"
	"github.com/NieR8/myProject/internal/env"
	"github.com/NieR8/myProject/orchestrator"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Загружаем конфигурацию
	config := env.LoadConfig()

	// Канал для остановки агента
	stop := make(chan struct{})

	// Контекст для остановки оркестратора
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем оркестратор
	orch := orchestrator.NewOrchestrator(config.OrchestratorAddr)
	go func() {
		log.Printf("Оркестратор запущен на %s", config.OrchestratorAddr)
		if err := orch.Run(ctx); err != nil {
			log.Printf("Ошибка оркестратора: %v", err)
		}
	}()

	// Запускаем агента
	agt := agent.NewAgent()
	go func() {
		log.Printf("Агент запущен с %d вычислителями", config.ComputingPower)
		agt.Run(stop)
	}()

	// Ожидаем сигнал остановки
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Останавливаем приложение
	log.Println("Получен сигнал остановки")
	close(stop) // Останавливаем агента
	cancel()    // Останавливаем оркестратор
	log.Println("Приложение остановлено")
}
