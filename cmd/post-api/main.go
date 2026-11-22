package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/config"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/database"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/handler"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/store"
)

func main() {
	// fmt.Println("Hello, World!")

	// what to do here?
	// 1. load config
	// 2. initialize logger
	// 3. setup router
	// 4. initialize storage
	// 5. initialize handlers
	// 6. run server

	// 1. load config
	cfg := config.MustLoad()

	// 2. initialize logger
	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Println("starting post-api")

	// 3. setup router
	router := http.NewServeMux()

	// 4. initialize storage
	db, dbErr := database.Open(cfg.StoragePath, logger)
	if dbErr != nil {
		logger.Fatalf("failed to open the database, error: %v", dbErr)
	}
	defer db.Close()

	migrateErr := database.Migrate(db, logger)
	if migrateErr != nil {
		logger.Fatalf("failed to migrate the database, error: %v", migrateErr)
	} else {
		logger.Println("database migrated successfully")
	}

	// 5. initialize handlers
	postStore := store.NewPostStore(db)
	postsHandler := handler.NewPostsHandler(postStore)
	handler.RoutesRegister(router, postsHandler)

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// 6. setup server
	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	logger.Println("server started on port", cfg.HTTPServer.Addr)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", err)
		}
	}()

	// get the signal from the channel if the server is interrupted
	<-done

	logger.Println("Server shutting down")

	// set a timeout for the shutdown for 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// cancel the context when the function returns so new requests are not accepted
	defer cancel()

	// ask the server to shutdown gracefully in case error occurs
	err := server.Shutdown(ctx)

	if err != nil {
		logger.Fatal("failed to shutdown server", err)
	} else {
		logger.Println("server shutdown gracefully")
	}

}
