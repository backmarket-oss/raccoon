.DEFAULT_GOAL := help

GOLANGCI_LINT_TIMEOUT ?= 1m

.PHONY: build
build: ## Build cli
	go build .

.PHONY: install
install: ## Install raccoon locally
	go install .

.PHONY: test
test: ## Launch go unit tests suite
	go test ./... -cover

.PHONY: lint
lint: ## Lint go code
	golangci-lint run -c .golangci.yaml --timeout $(GOLANGCI_LINT_TIMEOUT)

.PHONY: doc
doc: ## Launch godoc
	godoc -http=:6060

.PHONY: clean
clean: ## Cleanup
	go clean

.PHONY: dev
dev: build lint test ## Validate your code while developing


.PHONY: aws-login
aws-login: ## Sign in on AWS ECR
	aws ecr get-login-password --region us-east-1 | \
	  docker login --password-stdin -u AWS $(ECR_REPOSITORY)


.PHONY: docker-build
docker-build: ## Build docker image raccoon
	docker build --no-cache -t raccoon:latest .

.PHONY: docker-push
docker-push: ## Tag as latest & push last raccoon local images to AWS ECR
	@if [ -z "$$CIRCLE_TAG" ]; then \
		echo '$$CIRCLE_TAG env variable not set, aborting docker-push target execution'; \
	else \
		docker tag raccoon:latest ${ECR_REPOSITORY}/raccoon:latest; \
		docker tag raccoon:latest ${ECR_REPOSITORY}/raccoon:$$CIRCLE_TAG; \
		docker push ${ECR_REPOSITORY}/raccoon:latest; \
		docker push ${ECR_REPOSITORY}/raccoon:$$CIRCLE_TAG; \
	fi

.PHONY: help
help: ## Help gives you the list of available targets
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
