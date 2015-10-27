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
* command a query to multiple servers (server tree)
* distribute transaction logs
* manage a cluster-global unique key
* coordinate a distributed transaction
* abstract interface that hides an append-only immutable dataset from developers
* failover
* result set merge & flexible scoring
* connection pool with load balancing
* database operations without downtime

See also
* [MySQL Proxy](http://dev.mysql.com/doc/mysql-proxy/)
* [MariaDB MaxScale](https://mariadb.com/products/mariadb-maxscale)
* [Twitter Gizzard](https://github.com/twitter/gizzard)
* [LinkedIn Databus](https://engineering.linkedin.com/data-replication/open-sourcing-databus-linkedins-low-latency-change-data-capture-system) ([1](http://data.linkedin.com/projects/databus))
* [Google Vitess](https://github.com/youtube/vitess)
* [Tumblr Jetpants](https://github.com/tumblr/jetpants)
* [Netflix Dynomite](http://techblog.netflix.com/2014/11/introducing-dynomite.html)
* [AWS Database Migration Service](http://aws.amazon.com/dms/)
* [Flyway](http://flywaydb.org/) (database migration tool)
* [Liquibase](http://www.liquibase.org/) (database refactoring tool)
* [Compose's High Availability PostgreSQL service](https://blog.compose.io/high-availability-for-postgresql-batteries-not-included/)
* RDBMS vendor-specfic replication
