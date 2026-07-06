package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"renaper_mitramite/internal/bootstrap"
	"renaper_mitramite/internal/domain"
	"syscall"
)

func main() {
	// Корневой контекст процесса, отменяется по сигналам ОС
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	bootstrap.Run(ctx)

	// Ожидание завершения работы
	<-domain.Exit

	log.Println("Приложение renaper_mitramite завершено")
}
