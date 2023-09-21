package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/SerjZimmer/devops/internal/api"
	"github.com/SerjZimmer/devops/internal/config"
	"github.com/SerjZimmer/devops/internal/storage"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	server       *http.Server
	shutdownChan = make(chan struct{})
)

func main() {

	c := config.NewConfig()
	handler := api.NewHandler(storage.NewMetricsStorage())

	go func() {
		mRouter(handler)
		if err := run(c); err != nil {
			panic(err)
		}
	}()
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
		<-sigchan

		close(shutdownChan)
	}()

	<-shutdownChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при завершении работы сервера: %v\n", err)
	}

	os.Exit(0)
}

func run(c *config.Config) error {
	fmt.Printf("Сервер запущен на %v\n", c.Address)

	server = &http.Server{Addr: c.Address}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-shutdownChan
	fmt.Println("Завершение работы сервера...")

	// Завершаем работу сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при завершении работы сервера: %v\n", err)
	}

	return nil
}

func mRouter(handler *api.Handler) {
	r := mux.NewRouter()

	r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetric).Methods("POST")
	r.HandleFunc("/value/{metricType}/{metricName}", handler.GetMetric).Methods("GET")
	r.HandleFunc("/", handler.GetMetricsList).Methods("GET")
	http.Handle("/", r)
}
