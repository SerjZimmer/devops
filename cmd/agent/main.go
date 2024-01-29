package main

import (
	"time"

	config "github.com/SerjZimmer/devops/internal/config/agent"
	"github.com/SerjZimmer/devops/internal/storage"
)

// main является функцией точки входа в приложение.
// Здесь инициализируются конфигурация, хранилище метрик,
// а также запускаются горутины для периодического сбора и отправки данных.
func main() {

	printBuildInfo()

	c := config.New()
	s := storage.NewMetricsStorage(c.Storage)
	go func() {
		for {
			poll(s)
			time.Sleep(time.Second * time.Duration(c.PollInterval))
		}
	}()

	for {
		send(s, c)
		sendAllInBatches(s, c, 5)
		time.Sleep(time.Duration(c.ReportInterval) * time.Second)
	}

}
