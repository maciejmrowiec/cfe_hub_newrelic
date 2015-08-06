package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// build number set on during linking
var minversion string

type AppConfig struct {
	verbose     bool
	newRelicKey string
	interval    int
	version     bool
	filename    string
}

func HandleUserOptions() AppConfig {

	var config AppConfig

	flag.BoolVar(&config.verbose, "verbose", false, "Verbose mode")
	flag.StringVar(&config.newRelicKey, "key", "", "Newrelic license key (required)")
	flag.IntVar(&config.interval, "interval", 1, "Sampling interval [min]")
	flag.BoolVar(&config.version, "version", false, "Print version")
	flag.StringVar(&config.filename, "spec", "", "API request spec file. If not provided the API module will be disabled. (optional)")

	flag.Parse()

	if config.version {
		fmt.Printf("Build: %s\n", minversion)
		os.Exit(0)
	}

	if config.newRelicKey == "" {
		flag.PrintDefaults()
		log.Fatal("Required parameter missing.")
	}

	if config.filename != "" {
		if fileStat, err := os.Stat(config.filename); err != nil || fileStat.IsDir() {
			flag.PrintDefaults()
			log.Fatal("Failed to open spec file:" + err.Error())
		}
	}

	return config
}
