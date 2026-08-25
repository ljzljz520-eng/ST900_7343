package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"studio-console/config"
	"studio-console/httpapi"
	"studio-console/service"
	"studio-console/store"
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.FromEnv()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	database, err := store.Open(cfg.DatabasePath)
	if err != nil {
		return err
	}
	defer database.Close()
	manager, err := service.NewManager(database, systemClock{})
	if err != nil {
		return err
	}
	handler, err := httpapi.NewServer(database, manager, cfg.MaxBodyBytes)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: cfg.Address, Handler: handler, ReadHeaderTimeout: time.Duration(cfg.ReadTimeout) * time.Second, ReadTimeout: time.Duration(cfg.ReadTimeout) * time.Second, WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second}
	log.Printf("studio console listening on %s", cfg.Address)
	return server.ListenAndServe()
}
