# Validator statistics to Prometheus

This tool fetches validator statistics from the provided proxy, parses the response into Prometheus format and starts a web server where the metrics will be served through `/metrics` endpoint.

If no BLS key is provided on `config.toml` file, it will serve the metrics for all validators on the network.
