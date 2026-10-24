# -------------------------------------------------------------------------------------------
# VARIABLES: Variable declarations to be used within make to generate commands.
# -------------------------------------------------------------------------------------------
DEVELOP_DIR := development
COMPOSE := docker compose --project-directory ${DEVELOP_DIR} -f ${DEVELOP_DIR}/docker-compose.yml
ifeq ($(GITHUB_ACTIONS),true)
COMPOSE += --progress plain
endif
LIBRENMS_TOKEN ?= $(shell grep '^LIBRENMS_TOKEN=' ${DEVELOP_DIR}/.env | cut -d '=' -f2-)

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
	LIBRENMS_TOKEN=$(LIBRENMS_TOKEN) TF_ACC=1 go test -v -cover -timeout 120m ./internal/provider/

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
	$(COMPOSE) exec librenms lnms user:add -r admin -p Ad0cj3lasd0 admin
	$(COMPOSE) exec librenms sh -c 'DB="mariadb -h db -u $$MYSQL_USER -p$$MYSQL_PASSWORD -D librenms"; \
      if [ -n "$$($$DB -N -e "SHOW TABLES LIKE \"api_tokens\";")" ]; then \
        $$DB -e "INSERT INTO api_tokens (id,user_id,token_hash,description,disabled) VALUES (1,1,\"$$LIBRENMS_TOKEN\",\"\",0);"; \
      else \
        $$DB -e "INSERT INTO personal_access_tokens (tokenable_type,tokenable_id,name,token,abilities,created_at,updated_at) \
          VALUES (CONCAT(\"App\",CHAR(92),\"Models\",CHAR(92),\"User\"),1,\"ci\",SHA2(\"$$LIBRENMS_TOKEN\",256),JSON_ARRAY(\"*\"),NOW(),NOW());"; \
      fi'

.env:
	@if [ ! -f "${PWD}/${DEVELOP_DIR}/.env" ]; then \
	   echo "Error: Missing .env file"; \
	   exit 1; \
	fi

.PHONY: fmt lint test testacc build install generate dev-cli dev-debug dev-destroy dev-restart dev-start dev-stop dev-testacc .env
