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
}

func HandleUserOptions() AppConfig {

	var config AppConfig

	flag.BoolVar(&config.verbose, "verbose", false, "Verbose mode")
	flag.StringVar(&config.newRelicKey, "key", "", "Newrelic license key (required)")
	flag.IntVar(&config.interval, "interval", 1, "Sampling interval [min]")
	flag.BoolVar(&config.version, "version", false, "Print version")

	flag.Parse()

	if config.version {
		fmt.Printf("Build: %s\n", minversion)
		os.Exit(0)
	}

	if config.newRelicKey == "" {
		flag.PrintDefaults()
		log.Fatal("Required parameter missing.")
	}

	return config
}
