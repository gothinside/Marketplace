

COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

export GO111MODULE=on

.PHONY: build
build: 
	@echo "-- building binary"
	go build \
		-ldflags "-X main.buildHash=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
		-o ./bin/shopql \
		./cmd/shopql/server.go

all:
	go run ./bin/shopql

.PHONY: test
test:
	go test -v

.PHONY: install
install:
	go install github.com/99designs/gqlgen@v0.17.45

.PHONY: init
init:
	go run github.com/99designs/gqlgen init

.PHONY: gen
gen: 
	@echo "-- generatiog graphql files"
	go run github.com/99designs/gqlgen generate 

.PHONY: docker
docker: 
	@echo "-- building docker container"
	docker build -f build/Dockerfile -t shopql .

.PHONY: docker_run
docker_run: 
	@echo "-- starting docker container"
	docker run -it -p 8080:8080 shopql
	
.PHONY: docker_compose
docker_compose: 
	@echo "-- starting docker compose"
	docker compose -f ./deployment/docker-compose.yml up

.PHONY: dcb
dcb: 
	@echo "-- starting docker compose with build"
	docker compose -f ./deployment/docker-compose.yml up --build
# --verbose --config ./gqlgen.yml
