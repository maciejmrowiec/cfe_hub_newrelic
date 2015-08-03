package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	platform "github.com/yvasiyarov/newrelic_platform_go"
	"log"
	"os"
)

func main() {

	config := HandleUserOptions()

	// open database connection
	db, err := OpenDatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	plugin := platform.NewNewrelicPlugin(minversion, config.newRelicKey, 300)

	plugin.AddComponent(InitHubPerformanceStatsComponent(db, hostname, config.verbose))
	plugin.AddComponent(InitNetworkingStatsComponent(db, hostname, config.verbose))
	plugin.AddComponent(InitClientStatsComponent(db, hostname, config.verbose))
	plugin.AddComponent(InitMaintenanceStatsComponent(db, hostname, config.verbose))
	plugin.AddComponent(InitAPIStatsComponent(db, hostname, config.verbose))

	plugin.Verbose = config.verbose
	plugin.Run()
}

func OpenDatabaseConnection() (*sql.DB, error) {
	connectionUri := "postgres://root@localhost:5432/cfdb?sslmode=disable"
	log.Println("Connecting:", connectionUri)

	db, err := sql.Open("postgres", connectionUri)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db, err
}
