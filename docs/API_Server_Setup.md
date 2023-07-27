# Running API Server

This document explains how you can setup a development environment for `ILLA Builder` API server.

## Pre-requisites

- [Go](https://go.dev/doc/install)

- [PostgreSQL](https://www.postgresql.org/download/)

- [illa-supervisor-backend](https://github.com/illacloud/illa-supervisor-backend)

## Local Setup

1. Setup the PostgreSQL database

    - Running the [script](../scripts/postgres-init.sh) to create the database and tables

2. Setup the `illa-supervisor-backend`

    - Following the setup steps in [illa-supervisor-backend](https://github.com/illacloud/deploy-illa-manually/tree/main/build-by-yourself#build-illa-supervisor-backend)

3. Change the default env config

   Change the default env config in `pkg/db/connection.go` to the PostgreSQL config.

   Change the default env config in `internal/util/supervisior/token_validator.go` to the `illa-supervisor-backend` config.

4. Running the ILLA Builder API server

    ```bash
    go run github.com/illacloud/builder-backend/cmd/http-server
    ```

   This will start the ILLA Builder API server on  `http://127.0.0.1:8001`.

5. Extract the JWT token for the user `root`

    ```bash
    curl 'http://{{illa-supervisor-backend-addr}}/api/v1/auth/signin' --data-raw '{"email":"root","password":"password"}' -v
    ```

   Get the value of response header `illa-token` as the next API call's `Authorization` header value.

6. Test the API server

    ```bash
    curl 'http://127.0.0.1:8001/api/v1/teams/:teamID/apps' -H 'Authorization: {{Value of response header `illa-token`}}'
    ```

   The value of `:teamID` is `ILAfx4p1C7d0`.

## Need Assistance

- If you are unable to resolve any issue while doing the setup, please feel free to ask questions on our [Discord channel](https://discord.com/invite/illacloud) or initiate a [Github discussion](https://github.com/orgs/illacloud/discussions). We'll be happy to help you.
- In case you notice any discrepancy, please raise an issue on Github.