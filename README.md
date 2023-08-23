# crud-product-endpoints
crud-product-endpoints made with chi router and postgres

## Automating Program Compilation with a Makefile
- To build target use: make build-app
- To run target use: make run
- To run API inside docker container use: make docker

## Project environment variables
- HTTP_LISTEN_ADDRESS=:8080
- DB_HOST=
- DB_PORT=
- DB_USER=
- DB_PASSWORD=
- DB_NAME=
- DB_SSLMODE=

## Docker
### Installing postgres as a Docker container
- docker run --name postgresdb -e POSTGRES_PASSWORD=mysecretpassword -d -p 5432:5432 postgres:latest
