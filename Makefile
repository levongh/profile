GO               = go
M                = $(shell printf "\033[34;1m>>\033[0m")
GOBIN			 ?= $(PWD)/bin
TARGET_DIR       ?= $(PWD)/.build
MIGRATIONS_DIR	 = ./db/migrations/
TEST_STORAGE_DSN = 'postgres://postgres:postgres@localhost:5440/postgres_test?sslmode=disable&binary_parameters=yes'

.PHONY: all
all: build test

.PHONY: build
build: ## Build 'profile' binary
	$(info $(M) building profile...)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(TARGET_DIR)/profile ./cmd/*.go

watch: install-tools ; ## Run 'profile' binaries that rebuild themselves on changes
	$(info $(M) run...)
	@$(GOBIN)/refresh run -c .refresh.yml

.PHONY: fmt
fmt: ## Format code
	$(info $(M) running gofmt...)
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GO) fmt $$d/*.go || ret=$$? ; \
		done ; exit $$ret

.PHONY: install-tools
install-tools: $(GOBIN) ## Install tools needed for development
	@GOBIN=$(GOBIN) $(GO) install -mod=readonly -tags 'postgres' \
		github.com/markbates/refresh \
		github.com/golang-migrate/migrate/v4/cmd/migrate

.PHONY: lint
lint: install-tools ## Run linters
	$(info $(M) running linters...)
	golangci-lint run --timeout 5m0s ./...

.PHONY: lintfix
lintfix: install-tools ## Try to fix linter issues
	$(info $(M) fixing linter issues...)
	golangci-lint run --fix --verbose --timeout 2m0s ./... 2>&1 | \
		awk 'BEGIN{FS="="} /Fix/ { print $$3}' | \
		awk 'BEGIN{FS=","} {print " * ", $$1, $$2, $$8, $$9, $$10, $$11}' | \
		sed 's/\\"/"/g' | sed -e 's/&result.Issue{//g' | sed 's/token.Position//'

.PHONY: test
test: ## Run all tests
	$(info $(M) running tests...)
	@TEST_STORAGE_DSN=$(TEST_STORAGE_DSN) $(GO) test ./... -v -p=1 -cover

.PHONY: db-migrate
db-migrate: ## Run migrate command
	$(info $(M) running DB migrations...)
	@$(GOBIN)/migrate -path "$(MIGRATIONS_DIR)" -database "$(STORAGE_DSN)" $(filter-out $@,$(MAKECMDGOALS))

.PHONY: db-create-migration
db-create-migration: ## Create a new database migration file
	$(info $(M) creating DB migration...)
	@$(GOBIN)/migrate create -ext sql -dir "$(MIGRATIONS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

.PHONY: generate
generate: ## Run go generate
	$(info $(M) generating...)
	@$(GO) generate ./...

.PHONY: vet
vet: ## Run go vet
	$(info $(M) vetting source...)
	@go vet ./...

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(TARGET_DIR)

.PHONY: test-createdb
test-createdb:
	@docker-compose exec db createdb --username=postgres --owner=postgres postgres_test

.PHONY: $(GOBIN)
$(GOBIN):
	@mkdir -p $(GOBIN)

help:                   ##Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

%:
	@:

# For bitbucket-pipeline.yml
define run-tests
# 	docker-compose -f local-pipeline.yml run --rm nanos-profile /bin/bash -c "./profile
	docker save -o tmp-$(1)-image.docker ${DOCKER_REGISTRY}/nanos/$(2):$(1)
endef

define build_common
	docker build -t ${DOCKER_REGISTRY}/nanos/$(2):$(1) -f Dockerfile --build-arg TARGET_DIR=/app --build-arg SSH_PRIVATE_KEY="${SSH_PRIVATE_KEY}" --build-arg GOBIN=/.bin .
	docker tag ${DOCKER_REGISTRY}/nanos/$(2):$(1) profile_local_go:latest
	docker save -o tmp-$(1)-image.docker ${DOCKER_REGISTRY}/nanos/$(2):$(1)
endef

define push_common
	docker load -i tmp-$(1)-image.docker
	docker tag ${DOCKER_REGISTRY}/nanos/$(2):$(1) ${DOCKER_REGISTRY}/nanos/$(2):$(1)-${BITBUCKET_COMMIT}
	docker push ${DOCKER_REGISTRY}/nanos/$(2):$(1)
	docker push ${DOCKER_REGISTRY}/nanos/$(2):$(1)-${BITBUCKET_COMMIT}
endef

define deploy_common
    $(foreach i,$(services),aws ecs update-service --cluster nanos-$(1) --service nanos-$(i)-$(1) --force-new-deployment;)
endef

services = profile

test-app:
	$(call run-tests,${ENV_NAME},${ECR_NAME})

build-app:
	$(call build_common,${ENV_NAME},${ECR_NAME})

push-app:
	$(call push_common,${ENV_NAME},${ECR_NAME})

deploy-app:
	$(call deploy_common,${ENV_NAME})

slack-notification:
	curl -s -X POST ${SLACK_NOTIFICATION_URL} \
	-H "content-type:application/json" \
	-d '{"text":"[${BITBUCKET_REPO_SLUG}] [${BITBUCKET_BRANCH}] ${MESSAGE}"}'
