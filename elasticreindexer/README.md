# Elastic reindexer tool

This tool is able to reindex all mappings and data from an instance of Elasticsearch to another.

## How to use

- update the Elasticsearch instance configuration for both source and destination inside `config.toml`;
- update the indices list inside `config.toml`;
- run `go build`;
- run `./elasticreindexer` that will begin the process. Continuous prints should be observerd.

## Audience

This tool should be as generic as possible, and it doesn't have any custom code related to Elrond instances
of Elasticsearch
