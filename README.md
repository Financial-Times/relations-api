[![CircleCI](https://circleci.com/gh/Financial-Times/relations-api/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/relations-api/tree/master)
[![Coverage Status](https://coveralls.io/repos/github/Financial-Times/relations-api/badge.svg)](https://coveralls.io/github/Financial-Times/relations-api)

# Relations API

Relations API is an internally used API for retrieving content collection related content.
That is:
- content of CURATED relations
- content of CONTAINS relations for a given content or content collection (content package)

## Usage

### Install

Download the source code, dependencies and build the binary:

```shell script
go get github.com/Financial-Times/relations-api
cd $GOPATH/src/github.com/Financial-Times/relations-api
go install
```

### Tests

* Run unit tests only: `go test -race ./...`
* Run unit and integration tests:

    In order to execute the integration tests you must provide GITHUB_USERNAME and GITHUB_TOKEN values, because the service is depending on internal repositories.
    ```
    GITHUB_USERNAME=<username> GITHUB_TOKEN=<token> \
    docker-compose -f docker-compose-tests.yml up -d --build && \
    docker logs -f test-runner && \
    docker-compose -f docker-compose-tests.yml down -v
    ```

### Running locally

Run the binary (using the help flag to see the available optional arguments):

```shell script
$GOPATH/bin/relations-api [--help]
```

Options:

```shell script
--neo-url               neo-url value must use the bolt protocol (env $NEO_URL) (default "bolt://localhost:7687")
--port                  Port to listen on (env $PORT) (default "8080")
--cache-duration        Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '9900' seconds (env $CACHE_DURATION) (default "30s")
--api-yml               Location of the API Swagger YML file. (env $API_YML) (default "./api.yml")
--log-level             Logging level (DEBUG, INFO, WARN, ERROR) (env $LOG_LEVEL) (default "INFO")
--db-driver-log-level   Db's driver log level (DEBUG, INFO, WARN, ERROR) (env $DB_DRIVER_LOG_LEVEL) (default "WARN")
```



## Endpoints

### Application specific endpoints:

* /content/{uuid}/relations
* /contentcollection/{uuid}/relations

### Admin specific endpoints:

* /ping
* /build-info
* /__ping
* /__build-info
* /__health
* /__gtg

## Examples

#### For /content/{uuid}/relations endpoint:

`GET https://pre-prod-uk-up.ft.com/__relations-api/content/9b6eb364-0275-11e7-b9ac-52b4e2bf8289/relations`

```
{
       "curatedRelatedContent": [{
           "id": "http://api.ft.com/things/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90",
           "apiUrl": "http://api.ft.com/content/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90"
           }]
        "contains": [{
           "id": "http://api.ft.com/things/74bd05b4-edca-11e6-1234-ee7d9c5b3b90",
           "apiUrl": "http://api.ft.com/content/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90"
           },
           {
           "id": "http://api.ft.com/things/74bd05b4-edca-11e6-1313-ee7d9c5b3b90",
           "apiUrl": "http://api.ft.com/content/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90"
           }]
        "containedIn": [{
           "id": "http://api.ft.com/things/74bd05b4-adsd-1342-abbc-ee7d9c5b3b90",
           "apiUrl": "http://api.ft.com/content/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90"
           }]
   }
```

#### For /contentcollection/{uuid}/relations endpoint (for content package):

`GET https://pre-prod-uk-up.ft.com/__relations-api/content/9b6eb364-0275-11e7-b9ac-52b4e2bf8289/relations`

```
{
        "contains": [{
           "id": "http://api.ft.com/things/74bd05b4-edca-11e6-1234-ee7d9c5b3b90",
           "apiUrl": "http://api.ft.com/content/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90"
           },
           {
           "id": "http://api.ft.com/things/74bd05b4-edca-11e6-1313-ee7d9c5b3b90",
           "apiUrl": "http://api.ft.com/content/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90"
           }]
        "containedIn": [{
           "id": "http://api.ft.com/things/74bd05b4-adsd-1342-abbc-ee7d9c5b3b90",
           "apiUrl": "http://api.ft.com/content/74bd05b4-edca-11e6-abbc-ee7d9c5b3b90"
           }]
   }
```
