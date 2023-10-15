.PHONY: run
run:
	@go run cmd/swiss-knife/main.go

.PHONY: build
build:
	@echo "Building..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-musl-gcc go build -ldflags '-extldflags "-static"' -o bin/swiss-knife github.com/siriusfreak/swiss-knife/cmd/swiss-knife

.PHONY: image
image:
	@docker build -t registry.i.siriusfrk.ru/swiss-knife:latest -f deploy/Dockerfile .

.PHONY: push
push:
	@docker push registry.i.siriusfrk.ru/swiss-knife:latest

.PHONY: deploy
deploy:
	cd deploy && terraform init -backend-config=backend-config.tfvars
	cd deploy && terraform apply -var-file=backend-config.tfvars  -auto-approve

.PHONY: all
all: build image push deploy
