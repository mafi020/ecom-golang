package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mafi020/ecom-golang/internal/bootstrap"
)

func main() {
	app := bootstrap.InitializeApp()
	srv, err := app.RunHTTP()

	if err != nil {
		log.Fatal(err)
	}

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.ShutdownHTTP(srv); err != nil {
		log.Fatal(err)
	}
}
