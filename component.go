package main

import (
	"database/sql"
	platform "github.com/yvasiyarov/newrelic_platform_go"
)

func InitHubPerformanceStatsComponent(db *sql.DB, hostname string, verbose bool) platform.IComponent {

	component := platform.NewPluginComponent(hostname, "com.github.maciejmrowiec.cfe_hub_newrelic", verbose)

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

	// Count deltas and rebases
	component.AddMetrica(NewLocalCountDiagnostics(db, "consumer_processing_time_per_host", DELTA, 300))
	component.AddMetrica(NewLocalCountDiagnostics(db, "consumer_processing_time_per_host", REBASE, 300))

	component.AddMetrica(NewLocalCountDiagnostics(db, "duplicate_report", DELTA, 300))
	component.AddMetrica(NewLocalCountDiagnostics(db, "duplicate_report", REBASE, 300))

	// Pipeline measurements delta + rebase (total average)
	component.AddMetrica(NewLocalAverageDiagnostics(db, "consumer_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "hub_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "redis_processing_time_per_host", "", 300, "pipeline"))
	component.AddMetrica(NewLocalAverageDiagnostics(db, "redis_wait_time_per_host", "", 300, "pipeline"))

	// Estimated max hub capacity for cf-hub and cf-consumer
	component.AddMetrica(NewEstimatedCapacity("average/capacity/cf-hub", db, "hub", 300))
	component.AddMetrica(NewEstimatedCapacity("average/capacity/cf-consumer", db, "consumer", 300))

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
	component.AddMetrica(NewAverageBenchmark("hub/agent/maintenance_daily", 300, db, "cfe_internal_management_postgresql_vacuum:methods:hub"))
	component.AddMetrica(NewAverageBenchmark("hub/agent/maintenance_weekly", 300, db, "cfe_internal_management_postgresql_maintenance:methods:hub"))
	component.AddMetrica(NewAverageBenchmark("hub/agent/maintenance_report_history", 300, db, "cfe_internal_management_report_history:methods:hub"))

	return component
}

func InitAPIStatsComponent(db *sql.DB, hostname string, verbose bool) platform.IComponent {

	component := platform.NewPluginComponent(hostname, "com.github.maciejmrowiec.cfe_hub_newrelic", verbose)

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

	return component
}
