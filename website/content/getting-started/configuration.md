---
title: Configuration
weight: 2
---

Here's full list of config to run the Ngendika server:

```yaml
# settings for Transport layer
transport:
  http:
    port: 1234

# dependencies connection
# note that key must be alphanumeric only, e.g: db1, postgres1, mysql1
## define all database connection at once
databaseResources:
  db1:
    disable: false
    driver: "postgres"
    postgres:
      debug: true
      dsn: "user=postgres password=postgres host=localhost port=5433 dbname=ngendika sslmode=disable" # Data Source Name
  cfg1:
    disable: false
    driver: config

# settings each repository, select based on dependencies connection
## appstore to save application information
appService:
  dbLabel: db1 # refer to databaseResources

emailConfigService:
  dbLabel: db1 # refer to databaseResources

fcmService:
  dbLabel: db1 # refer to databaseResources

## message service will connect to via function call:
## * appService
## * emailConfigService
## * fcmService
msgService:
  dbLabel: db1 # refer to databaseResources
  maxParallel: 10 # number of semaphore to limit the number of goroutines working on parallel tasks

```

## How to read the configuration file

* First you need to specify all the database connection under `databaseResources`, it uses key as the label references.
  For example, you define the `databaseResources` something like this:

```yaml
databaseResources:
  db1:
    disable: false
    driver: "postgres"
    postgres:
    debug: true
    dsn: "user=postgres password=postgres host=localhost port=5433 dbname=ngendika sslmode=disable" # Data Source Name
```

  Then you can use `db1` as the label references to the services.
  Using this approach, we try to separate the Database(s) and Service(s) into separated object.
  That means, we want to make it easier to move specific data into large resource when needed.
  For example, in future we may want to drop any unused data in Email config, it should not bother the App database.
  
* In services, such as `appService` or `msgService` we can define the database connection by referring it to the database connection label.