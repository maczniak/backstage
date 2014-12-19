backstage
=========

flexible database middleware toolkit for dirty jobs

**NOTE: This is a proof of concept, and does not have working codes yet.**

Feature
* modify datatabase query and query result
* manipulate transaction logs
* manage heterogeneous database cluster

Application
* partition (sharding)
* standby replica
* command a query to multiple servers
* distribute transaction logs
* manage a cluster-global unique key
* coordinate a distributed transaction
* abstract interface that hides an append-only immutable dataset from developers
* failover

See also
* [MySQL Proxy](http://dev.mysql.com/doc/mysql-proxy/)
* [Twitter Gizzard](https://github.com/twitter/gizzard)
* [LinkedIn Databus](https://engineering.linkedin.com/data-replication/open-sourcing-databus-linkedins-low-latency-change-data-capture-system) ([1](http://data.linkedin.com/projects/databus))
* RDBMS vendor-specfic replication
