ENV_NAME="route-256-ws-6"

build-all:
	cd cart && GOOS=linux GOARCH=amd64 make build
	cd loms && GOOS=linux GOARCH=amd64 make build
	cd notifier && GOOS=linux GOARCH=amd64 make build

run-all: build-all
	docker-compose up --force-recreate --build

compose-up: build-all
	docker-compose -p ${ENV_NAME} up -d --build --force-recreate

compose-down:
	docker-compose -p ${ENV_NAME} stop

compose-rm:
	docker-compose -p ${ENV_NAME} rm -fvs