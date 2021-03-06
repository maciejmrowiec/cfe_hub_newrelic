package main

import (
	"database/sql"
	"encoding/json"
	diag "github.com/maciejmrowiec/cfe_hub_newrelic/diagnostics"
	platform "github.com/yvasiyarov/newrelic_platform_go"
	"io/ioutil"
	"log"
)

func InitHubPerformanceStatsComponent(db *sql.DB, hostname string, verbose bool) platform.IComponent {

	component := platform.NewPluginComponent(hostname, "com.github.maciejmrowiec.cfe_hub_newrelic", verbose)

	// performane per delta and rebase
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "consumer_processing_time_per_host", diag.DELTA, 300, "byquery"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "consumer_processing_time_per_host", diag.REBASE, 300, "byquery"))

	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "hub_processing_time_per_host", diag.DELTA, 300, "byquery"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "hub_processing_time_per_host", diag.REBASE, 300, "byquery"))

	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "recivied_data_size_per_host", diag.DELTA, 300, "byquery"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "recivied_data_size_per_host", diag.REBASE, 300, "byquery"))

	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "redis_processing_time_per_host", diag.DELTA, 300, "byquery"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "redis_processing_time_per_host", diag.REBASE, 300, "byquery"))

	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "hub_collection_total_time", "", 300, "byquery"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "redis_wait_time_per_host", "", 300, "byquery"))

	// Count deltas and rebases
	component.AddMetrica(diag.NewLocalCountDiagnostics(db, "consumer_processing_time_per_host", diag.DELTA, 300))
	component.AddMetrica(diag.NewLocalCountDiagnostics(db, "consumer_processing_time_per_host", diag.REBASE, 300))

	component.AddMetrica(diag.NewLocalCountDiagnostics(db, "duplicate_report", diag.DELTA, 300))
	component.AddMetrica(diag.NewLocalCountDiagnostics(db, "duplicate_report", diag.REBASE, 300))

	// Pipeline measurements delta + rebase (total average)
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "consumer_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "hub_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "redis_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(diag.NewLocalAverageDiagnostics(db, "redis_wait_time_per_host", "", 300, "pipeline"))

	// Estimated max hub capacity for cf-hub and cf-consumer
	component.AddMetrica(diag.NewEstimatedCapacity("average/capacity/cf-hub", db, "hub", 300))
	component.AddMetrica(diag.NewEstimatedCapacity("average/capacity/cf-consumer", db, "consumer", 300))

	// Host count
	component.AddMetrica(NewHostCount("host/count", db))

	return component
}

func InitNetworkingStatsComponent(db *sql.DB, hostname string, verbose bool) platform.IComponent {

	component := platform.NewPluginComponent(hostname, "com.github.maciejmrowiec.cfe_hub_newrelic", verbose)

	// Lasteen incomming vs outgoing
	component.AddMetrica(NewConnectionEstablished("network/connections/count/incoming", db, "INCOMING", 300))
	component.AddMetrica(NewConnectionEstablished("network/connections/count/outgoing", db, "OUTGOING", 300))

	// Hub connection errors encountered by cf-hub (count)
	component.AddMetrica(NewConnectionErrorCount("network/error/count/ServerNoReply", db, "ServerNoReply", 300))
	component.AddMetrica(NewConnectionErrorCount("network/error/count/ServerAuthenticationError", db, "ServerAuthenticationError", 300))
	component.AddMetrica(NewConnectionErrorCount("network/error/count/InvalidData", db, "InvalidData", 300))
	component.AddMetrica(NewConnectionErrorCount("network/error/count/HostKeyMismatch", db, "HostKeyMismatch", 300))

	return component
}

func InitClientStatsComponent(db *sql.DB, hostname string, verbose bool) platform.IComponent {

	component := platform.NewPluginComponent(hostname, "com.github.maciejmrowiec.cfe_hub_newrelic", verbose)

	// Avg agent execution time per promises.cf / update.cf / failsafe.cf
	component.AddMetrica(NewAverageBenchmark("host/agent/avg_execution_failsafe.cf", 300, db, "CFEngine Execution (policy filename: '/var/cfengine/inputs/failsafe.cf')"))
	component.AddMetrica(NewAverageBenchmark("host/agent/avg_execution_update.cf", 300, db, "CFEngine Execution (policy filename: '/var/cfengine/inputs/update.cf')"))
	component.AddMetrica(NewAverageBenchmark("host/agent/avg_execution_promises.cf", 300, db, "CFEngine Execution (policy filename: '/var/cfengine/inputs/promises.cf')"))

	return component
}

func InitMaintenanceStatsComponent(db *sql.DB, hostname string, verbose bool) platform.IComponent {

	component := platform.NewPluginComponent(hostname, "com.github.maciejmrowiec.cfe_hub_newrelic", verbose)

	// Maintenance execution policy
	component.AddMetrica(NewAverageBenchmark("hub/maintenance/vacuum", 600, db, "cfe_internal_management_postgresql_vacuum:methods:hub"))
	component.AddMetrica(NewAverageBenchmark("hub/maintenance/vacuum_full", 600, db, "cfe_internal_management_postgresql_maintenance:methods:hub"))
	component.AddMetrica(NewAverageBenchmark("hub/maintenance/report_cleanups", 300, db, "cfe_internal_management_report_history:methods:hub"))

	return component
}

func InitAPIStatsComponent(db *sql.DB, hostname string, verbose bool, filename string) platform.IComponent {

	component := platform.NewPluginComponent(hostname, "com.github.maciejmrowiec.cfe_hub_newrelic", verbose)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return component
	}

	var list []Request
	if err := json.Unmarshal(data, &list); err != nil {
		log.Fatalln(err)
		return component
	}

	for _, request := range list {
		component.AddMetrica(NewApiTiming(request, "api/reponse/", "admin", "admin"))
	}

	return component
}
