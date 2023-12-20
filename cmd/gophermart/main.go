package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/The-Gleb/loyalty-system/internal/app"
	"github.com/The-Gleb/loyalty-system/internal/handlers"
	"github.com/The-Gleb/loyalty-system/internal/server"
	"github.com/The-Gleb/loyalty-system/internal/storage/database"
)

// TODO`s
// + Implement storage layer
// - Write unit tests
// - Implement error handling
// - Implement logger
// - Implement auth middleware
// - Impement Luhn algorithm

func main() {

	config := NewConfigFromFlags()

	db, err := database.ConnectDB(config.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}

	app := app.NewApp(db, config.AccrualAddress)

	handlers := handlers.New(app)

	s := server.New(config.RunAddress, handlers)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ServerShutdownSignal := make(chan os.Signal, 1)
		signal.Notify(ServerShutdownSignal, syscall.SIGINT)
		<-ServerShutdownSignal
		s.Shutdown(context.Background())
	}()

	err = server.Run(s)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	wg.Wait()
}
