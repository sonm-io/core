 ## Prometheus + Graphana example
 
 Follow these steps to set up Prometheus and Graphana metrics monitoring:
 
 * Execute `cd ./examples/prometheus_graphana && docker-compose up` to run Prometheus and Graphana;
 * Run the SONM components that you need;
 * Go to `http://localhost:3000`, user `admin`, password `pass`;
 * Create a new data source of type `Prometheus` that points to `http://localhost:9090`;
 * Create a new dashboard, then `Add Panel > Graph` and change panel's data source to the one you created at previous step.
 
 Now you can run queries and monitor components' state.
 
 Refer to [go-grpc-prometheus documentation](https://github.com/grpc-ecosystem/go-grpc-prometheus#useful-query-examples) for gRPC-related queries.
 
 The following custom metrics have been added:
 
 * `sonm_deals_current` -- a gauge that keeps track of current deals (only registered by Worker);
 * `sonm_tasks_current` -- a gauge that keeps track of current tasks (only registered by Worker);
 * `grpc_server_connections_current` -- a gauge that keeps track of current server connections.
 
 N.B.: you may have difficulties running this guide on a non-Linux host, consider using a VM if needed.