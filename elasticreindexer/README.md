# Elastic reindexer tool

- This tool can be used to copy an Elasticsearch  index/indices from a cluster to another. It can copy also the indices 
mappings ( we recommend to use it without copying the mappings ). The indices mappings part can be done with the `indices-creator` tool.
Also, with the `indices-creator` tool some properties of the indices can be changed ( depending on the needs ).

- The main scope of these tools is to can copy all the information from an Elasticsearch cluster to another ( all the indices that was
populated by the Elrond nodes from the genesis till the current time ).

## How to use
### STEP 1
- In order to can reindex all the information from an Elasticsearch cluster to another had to create all the indices mappings. 
To do that have to use the `indices-creator` tool:

    ```
    cd elasticreindexer
    cd cmd/indices-creator/
    go build 
    ```
  
- After the build is done have to update the `config/cluster.toml` file with the information about the Elasticsearch cluster. In the `cluster.toml` have to set the URL 
of the Elasticsearch cluster, and for what indices have to create the mappings.

- Optionally also the mappings can be customized ( open file with mappings for every index and set different settings, based on the needs)

- Run `./indices-creator` in order to create all the indices and mappings


_*This step can be skipped for the clusters the already have information indexed_ 

### STEP 2
- After the mappings and indices was created we can start to copy all the information. In order to do this have to use the `elasticreindexer` tool that 
will copy every index one by one.

    ```
    cd elasticreindexer
    cd cmd/elasticreindexer/
    go build 
    ```
- Update the Elasticsearch instance configuration for both source and destination inside `config.toml`. In the `config.toml` file have to set the 
URL of the `input` Elasticsearch instance ( the one from where have to copy all the information) and the `output` instance ( the one where all the information 
from the `input` instance will be copied )
- The `config.toml` file contains by default all the Elasticsearch indices that are populated by an Elrond observing-squad

- Also, if you want to copy indices with timestamp have to set in the `config.toml` file the `blockchain-start-time` ( by default is the one from the mainnet)

- Run `./elasticreindexer --skip-mappings` ( will start to reindex all the information from the input cluster in the output cluster based on the `config.toml` file)



## Audience

This tool should be as generic as possible, and it doesn't have any custom code related to Elrond instances
of Elasticsearch
