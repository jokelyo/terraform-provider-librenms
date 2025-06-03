# -------------------------------------------------------------------------------------------
# VARIABLES: Variable declarations to be used within make to generate commands.
# -------------------------------------------------------------------------------------------
DEVELOP_DIR := development
COMPOSE := docker compose --project-directory ${DEVELOP_DIR} -f ${DEVELOP_DIR}/docker-compose.yml

default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

dev-cli: .env
ifeq (,$(findstring librenms,$($COMPOSE ps --services --filter status=running)))
	@make dev-start
endif
	@$(COMPOSE) exec librenms bash

dev-debug: .env
	$(COMPOSE) up

dev-destroy:
	$(COMPOSE) down --volumes --remove-orphans

dev-restart: .env
	$(COMPOSE) restart

dev-start: .env
	$(COMPOSE) up -d

dev-stop:
	$(COMPOSE) down

# This target is meant to be used in a CI/CD pipeline as a single-use admin/api token setup.
dev-testacc: dev-start
	sleep 20
	$(COMPOSE) exec librenms lnms user:add -r admin -p admin admin
	$(COMPOSE) exec librenms sh -c 'mariadb -h db -u $$MYSQL_USER -p$$MYSQL_PASSWORD -D librenms \
      -e "insert into api_tokens (id,user_id,token_hash,description,disabled) VALUES (1,1,\"$$LIBRENMS_TOKEN\",\"\",0);"'

.env:
	@if [ ! -f "${PWD}/${DEVELOP_DIR}/.env" ]; then \
	   echo "Error: Missing .env file"; \
	   exit 1; \
	fi

.PHONY: fmt lint test testacc build install generate dev-cli dev-debug dev-destroy dev-restart dev-start dev-stop dev-testacc .env
