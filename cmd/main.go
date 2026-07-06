package main

import (
	"context"
	"os"
	"os/signal"
	"renaper_mitramite/internal/bootstrap"
	"renaper_mitramite/internal/domain"
	"syscall"

	"github.com/ikermy/AiR_Logger/v2/pkg/logger"
)

func main() {
	logger.StdOut().WithLogLevel(logger.DEBUG).Apply()

	// Корневой контекст процесса, отменяется по сигналам ОС
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	bootstrap.Run(ctx)

	// Ожидание завершения работы
	<-domain.Exit

	logger.Infoln("Приложение renaper_mitramite завершено")
}
