[![Build Status](https://drone.io/github.com/maciejmrowiec/cfe_hub_newrelic/status.png)](https://drone.io/github.com/maciejmrowiec/cfe_hub_newrelic/latest)

# cfe_hub_newrelic
New Relic plugin for monitoring CFEngine Hub 

```
go get github.com/lib/pq
go get github.com/yvasiyarov/newrelic_platform_go
go get github.com/maciejmrowiec/cfe_hub_newrelic
go build
```

To enable pg cleanup job monitoring please adjust policy (cfe_internal/enterprise/main.cf):

```
"hub" usebundle => cfe_internal_hub_maintain,
       handle => "cfe_internal_management_hub_maintain",
+      # action => measure_promise_time("cfe_internal_management_report_history"),
       comment => "Start the hub maintenance process";
```
