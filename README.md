# CSV to CRM uploader

A demonstration of reading customer data from a CSV file, saving to a Postgresql database, and uploading (POST) that to a CRM service provider.

## Getting Started

### Prerequesites:
- Git
- Go > 1.11.4
- [Docker](https://www.docker.com/) (Docker Engine > 1.13.0)
- docker-compose (A version that supports compose version 3.1)
- At least two terminal windows to watch the `crmIntegrator` as the `csvReader` reads and signals when customers are ready to be uploaded.
- [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) (to run the Go tests)

### Getting Going:
Run the following commands:
```
$ git clone https://github.com/dbyington/csv-crm-upload.git
$ cd csv-crm-upload
$ ./bin/setup.sh
$ ./bin/start.sh
```
This gets the source then builds the `csvReader` and `crmIntegrator` binaries and starts the postgres and mock CRM service containers. It may take a minute or two for the postgres server to start.
When `docker-compose` shows that the postgres service is healthy it is ready to go.

You'll see this;
```
$ docker-compose ps
          Name                         Command                 State               Ports
-------------------------------------------------------------------------------------------------
csv-crm-upload_crm_1        go run /root/crm/server.go      Up             0.0.0.0:8089->80/tcp
csv-crm-upload_postgres_1   docker-entrypoint.sh postgres   Up (healthy)   0.0.0.0:5432->5432/tcp
```

Now start the `crmIntegrator` service, in one terminal window:
```
$ ./crmIntegrator
```
This will begin outputting that it is checking for work. As long as there is no work or the CRM Service provider is unresponsive the interval between checks will continually grow larger (it increments using a fibonacci numbering scheme).

Then, to begin sending CSV data, in another terminal window, source the `.env` file and execute the `csvReader` with one of the supplied files located in the `assets` directory.
```
$ set -a
$ source .env
$ ./csvReader -f assets/MOCK_DATA.csv
```
You should see the `crmIntegrator` begin to stream log message about the CRM service responding with 503 errors. This is intentional. To simulate a flaky service provider the "mock" service responds to at most 10% of all requests with a 503 error.

Something like:
```
$ ./csvReader -filename=assets/MOCK_DATA.csv
2019/08/25 16:23:52 database open
2019/08/25 16:23:52 csv file open
2019/08/25 16:23:52 starting...
2019/08/25 16:23:52 Processing 5 customers
2019/08/25 16:23:52 done.
2019/08/25 16:23:52 Processing 10 customers
2019/08/25 16:23:52 done.
2019/08/25 16:23:52 Processing 15 customers
2019/08/25 16:23:52 done.
2019/08/25 16:23:52 Processing 31 customers
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 done.
2019/08/25 16:23:52 Processing 146 customers
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 post to CRM failed with (503) 503 Service Unavailable
2019/08/25 16:23:52 done.
```

If you run the exact same command you'll get a spew of errors because customers in the database must be unique. To rerun the command first clear the data in the database by running the command:
```
$ ./bin/refresh-db.sh
```
This will delete all rows from the database so you can start fresh. If you want to rebuild the database use the supplied `bin/wipe-db.sh` command.

To manually inspect the database content (requires SQL knowledge) you can connect to the database by running:
```
$ docker-compose run --rm postgres psql -h postgres -U csvcrm crm
```
And using the password `csvcrm`.

You can follow the mock CRM Service log output in another terminal window with:
```
$ docker-compose logs -f crm
```

### Running tests:
To execute the unit tests you can run `go test ./...` or run the helper script:
```
$ ./bin/test.sh
```

## Background
### Rules:
1. The CSV reading and CRM upload with be done by separate services
1. The CSV file will be of unknown length and must not be loaded, entirely into memory
1. Each customer's data must only be uploaded to the CRM service once
1. All service code will be written in Go using the standard Go library

### Additional notes:
- Steps need to be taken to handle failures in the CRM service provider
- Both CSV reading and CRM upload binaries will run on the same machine
