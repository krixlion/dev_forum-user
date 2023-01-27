#!make
include .env
export $(shell sed 's/=.*//' .env)

kubernetes = kubectl -n dev
docker-compose = docker compose -f docker-compose.dev.yml --env-file .env

mod-init:
	go mod init	github.com/krixlion/$(PROJECT_NAME)-$(AGGREGATE_ID)
	go mod tidy
	go mod vendor

run-local:
	go run ./...

build-local:
	go build

push-image: # param: version
	docker build deployment/ -t krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
	docker push krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)

# ------------- Kubernetes -------------

k8s-mount-project:
	mkdir -p /mnt/wsl/k8s-mount/${AGGREGATE_ID} && sudo mount --bind . /mnt/wsl/k8s-mount/${AGGREGATE_ID}

k8s-unit-test: # param: args
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -short -race ${args} ./...  

k8s-integration-test: # param: args
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -race ${args} ./...  

k8s-test-gen-coverage:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -coverprofile  cover.out ./...
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go tool cover -html cover.out -o cover.html

k8s-run-dev:
	- $(kubernetes) delete -R -f deployment/k8s/dev/resources/
	$(kubernetes) apply -R -f deployment/k8s/dev/resources/

# k8s-setup-tools:
# 	kubectl apply -f deployment/k8s/dev/dev-namespace.yml
# 	kubectl apply -R -f deployment/k8s/dev/kubernetes-dashboard.yml
# 	kubectl apply -R -f deployment/k8s/dev/metrics-server.yml

# k8s-setup-telemetry:
# 	$(kubernetes) apply -R -f deployment/k8s/dev/instrumentation/