package shared

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CatchShutdown возвращает контекст, который отменяется
// при получении SIGINT / SIGTERM / SIGQUIT.
func CatchShutdown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigCh
		cancel()
		<-sigCh
		os.Exit(1)
	}()

	return ctx
}
