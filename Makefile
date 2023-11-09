#!make
include .env
export $(shell sed 's/=.*//' .env)

kubernetes = kubectl -n dev
overlays-path = deployment/k8s/overlays

mod-init:
	go mod init	github.com/krixlion/$(PROJECT_NAME)-$(AGGREGATE_ID)
	go mod tidy
	go mod vendor

grpc-gen:
	docker run --rm -v $(shell pwd):/app --env-file .env krixlion/go-grpc-gen:${GO_VERSION}

build-image: # param: version
	# OTEL_EXPORTER_OTLP_ENDPOINT variable imported from '.env' file triggers tracing in docker and causes issues.
	OTEL_EXPORTER_OTLP_ENDPOINT="" docker build . -f deployment/Dockerfile -t krixlion/$(PROJECT_NAME)-$(AGGREGATE_ID):$(version)

push-image: build-image # param: version
	docker push krixlion/$(PROJECT_NAME)-$(AGGREGATE_ID):$(version)

# ------------- Kubernetes -------------

k8s-mount-project:
	mkdir -p /mnt/wsl/k8s-mount/${AGGREGATE_ID} && sudo mount --bind $(shell pwd) /mnt/wsl/k8s-mount/${AGGREGATE_ID}

k8s-db-migrate-up:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go run cmd/migrate/up/main.go

k8s-db-migrate-down:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go run cmd/migrate/down/main.go

k8s-unit-test: # param: args
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -short -race ${args} ./...  

k8s-integration-test: # param: args
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -race ${args} ./...  

k8s-test-gen-coverage:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -coverprofile  cover.out ./...
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go tool cover -html cover.out -o cover.html

k8s-run-dev: k8s-stop-dev
	$(kubernetes) -k $(overlays-path)/dev apply

k8s-stop-dev:
	- $(kubernetes) -k $(overlays-path)/dev delete 
