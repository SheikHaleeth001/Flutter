package main

//go:generate pipenv run python ./scripts/generate/generate.py

import (
	"ZenitusHTML/configuration"
	"ZenitusHTML/constants"
	"ZenitusHTML/router"
	"ZenitusHTML/utils"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const MaxHeaderBytes int = 1 << 20

func main() {
	fmt.Println("aa")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configPath := flag.String("conf", "", "configuration folder path give without trailing slash")
	flag.Parse()
	fmt.Println("a")
	if flag.NFlag() != 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	fmt.Println("b")

	filePath := strings.TrimSpace(*configPath) + "/config.json"
	schemaPath := strings.TrimSpace(*configPath) + "/schema.json"

	// config init
	cfg := configuration.NewViperConfig()
	config, err := cfg.Init(filePath, schemaPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Determine port for HTTP service.
	port := config.GetString(utils.APP_PORT)
	if port == "" {
		port = "9090"
		log.Printf("defaulting to port %s", port)
	}

	mux, closeMux := router.NewChiRouter(cfg).InitRouter()

	// Start HTTP server.
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.GetInt(constants.SrvReadTimeOut)) * time.Second,
		WriteTimeout:   time.Duration(config.GetInt(constants.SrvWriteTimeout)) * time.Second,
		IdleTimeout:    time.Duration(config.GetInt(constants.SrvMaxIdleTimeOut)) * time.Second,
		MaxHeaderBytes: MaxHeaderBytes,
	}

	server.RegisterOnShutdown(closeMux)
	server.RegisterOnShutdown(config.Close)

	// Graceful shut down of server
	graceful := make(chan os.Signal, 1)
	signal.Notify(graceful, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-graceful
		log.Println("Shutting down server...")

		timeout := time.Duration(config.GetInt(constants.SrvTimeOut)) * time.Second
		ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
		defer cancelFunc()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not do graceful shutdown: %v\n", err)
		}
	}()

	log.Println("Listening server on ", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Not able to start server on port: %v error: %s", config.GetString(constants.Port), err)
	}

	log.Println("Server gracefully stopped...")
}
