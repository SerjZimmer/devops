package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/SerjZimmer/devops/internal/api"
	"github.com/SerjZimmer/devops/internal/config"
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

	config.FlagInit()

	go func() {
		mRouter()
		if err := run(); err != nil {
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

func run() error {
	fmt.Printf("Сервер запущен на %v\n", config.Address)

	server = &http.Server{Addr: config.Address}
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

func mRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", api.UpdateMetric).Methods("POST")
	r.HandleFunc("/value/{metricType}/{metricName}", api.GetMetric).Methods("GET")
	r.HandleFunc("/", api.GetMetricsList).Methods("GET")
	http.Handle("/", r)
}
