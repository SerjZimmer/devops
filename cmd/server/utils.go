package main

import (
	"fmt"
	"net/http"
)

var (
	server       *http.Server
	shutdownChan = make(chan struct{})
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	fmt.Println("Build version:", getOrDefault(buildVersion, "N/A"))
	fmt.Println("Build date:", getOrDefault(buildDate, "N/A"))
	fmt.Println("Build commit:", getOrDefault(buildCommit, "N/A"))
}

func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
