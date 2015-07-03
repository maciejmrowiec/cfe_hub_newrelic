package main

import (
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"github.com/yvasiyarov/newrelic_platform_go"
	"log"
	"os"
)

func main() {

	var verbose bool
	var newrelic_key string
	flag.StringVar(&newrelic_key, "key", "", "Newrelic license key")
	flag.BoolVar(&verbose, "v", false, "Verbose mode")

	flag.Parse()

	if newrelic_key == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// open database connection
	db, err := sql.Open("postgres", "postgres://root@localhost:5432/cfdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// regiter components
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	component := newrelic_platform_go.NewPluginComponent("hub/"+hostname, "com.github.maciejmrowiec.cfe_hub_newrelic")

	plugin := newrelic_platform_go.NewNewrelicPlugin("0.0.1", newrelic_key, 300)
	plugin.AddComponent(component)

	// performane per delta and rebase
	component.AddMetrica(NewLocalAverageDiagnostics(db, "consumer_processing_time_per_host", DELTA, 300, "byquery"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "consumer_processing_time_per_host", REBASE, 300, "byquery"))

	component.AddMetrica(NewLocalAverageDiagnostics(db, "hub_processing_time_per_host", DELTA, 300, "byquery"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "hub_processing_time_per_host", REBASE, 300, "byquery"))

	component.AddMetrica(NewLocalAverageDiagnostics(db, "recivied_data_size_per_host", DELTA, 300, "byquery"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "recivied_data_size_per_host", REBASE, 300, "byquery"))

	component.AddMetrica(NewLocalAverageDiagnostics(db, "redis_processing_time_per_host", DELTA, 300, "byquery"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "redis_processing_time_per_host", REBASE, 300, "byquery"))

	component.AddMetrica(NewLocalAverageDiagnostics(db, "hub_collection_total_time", "", 300, "byquery"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "redis_wait_time_per_host", "", 300, "byquery"))

	component.AddMetrica(NewLocalAverageDiagnostics(db, "duplicate_report", DELTA, 300, "byquery"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "duplicate_report", REBASE, 300, "byquery"))

	// Count deltas and rebases
	component.AddMetrica(NewLocalCountDiagnostics(db, "consumer_processing_time_per_host", DELTA, 300))
	component.AddMetrica(NewLocalCountDiagnostics(db, "consumer_processing_time_per_host", REBASE, 300))

	// Pipeline measurements delta + rebase (total average)
	component.AddMetrica(NewLocalAverageDiagnostics(db, "consumer_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "hub_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "redis_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "redis_wait_time_per_host", "", 300, "pipeline"))

	// Hub connection errors encountered by cf-hub (count)
	component.AddMetrica(NewConnectionErrorCount("network/error/count/ServerNoReply", db, "ServerNoReply", 300))
	component.AddMetrica(NewConnectionErrorCount("network/error/count/ServerAuthenticationError", db, "ServerAuthenticationError", 300))
	component.AddMetrica(NewConnectionErrorCount("network/error/count/InvalidData", db, "InvalidData", 300))

	// Avg agent execution time per promises.cf / update.cf / failsafe.cf
	component.AddMetrica(NewAverageBenchmark("host/agent/avg_execution_failsafe.cf", 300, db, "CFEngine Execution (policy filename: '/var/cfengine/inputs/failsafe.cf')"))
	component.AddMetrica(NewAverageBenchmark("host/agent/avg_execution_update.cf", 300, db, "CFEngine Execution (policy filename: '/var/cfengine/inputs/update.cf')"))
	component.AddMetrica(NewAverageBenchmark("host/agent/avg_execution_promises.cf", 300, db, "CFEngine Execution (policy filename: '/var/cfengine/inputs/promises.cf')"))

	// Lasteen incomming vs outgoing
	component.AddMetrica(NewConnectionEstablished("network/connections/count/incoming", db, "INCOMING", 300))
	component.AddMetrica(NewConnectionEstablished("network/connections/count/outgoing", db, "OUTGOING", 300))

	// Estimated max hub capacity for cf-hub and cf-consumer
	component.AddMetrica(NewEstimatedCapacity("average/capacity/cf-hub", db, "hub", 300))
	component.AddMetrica(NewEstimatedCapacity("average/capacity/cf-consumer", db, "consumer", 300))

	// Host count
	component.AddMetrica(NewHostCount("host/count", db))

	// query api tests
	// software updates trigger
	software_updates_trigger := &QueryTiming{
		api_call: QueryApi{
			User:     AdminUserName,
			Password: AdminPassword,
			BaseUrl:  BaseUrl,
			Resource: Query{
				Query: "SELECT count (*) AS failhost FROM (SELECT DISTINCT s_up.hostkey FROM softwareupdates s_up WHERE patchreporttype = 'AVAILABLE') AS c_query",
			},
		},
		name: "software_updates/trigger",
	}
	component.AddMetrica(software_updates_trigger)

	// software updates alert page
	software_updates_alert := &QueryTiming{
		api_call: QueryApi{
			User:     AdminUserName,
			Password: AdminPassword,
			BaseUrl:  BaseUrl,
			Resource: Query{
				Query:           `SELECT h.hostkey, h.hostname, count (s.patchname ) AS "c" FROM hosts h INNER JOIN softwareupdates s ON s.hostkey = h.hostkey WHERE patchreporttype = 'AVAILABLE' GROUP BY h.hostkey, h.hostname ORDER BY c DESC`,
				PaginationLimit: 50,
			},
		},
		name: "software_updates/alert",
	}
	component.AddMetrica(software_updates_alert)

	plugin.Verbose = verbose
	plugin.Run()
}
