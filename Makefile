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

k8s-db-migrate-up:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go run cmd/migrate/up/main.go

k8s-db-migrate-down:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go run cmd/migrate/down/main.go

k8s-db-seed:
	$(kubernetes) exec -it cockroachdb-cluster-0 -- cockroach sql --insecure --execute='TRUNCATE "users";'
	$(kubernetes) exec -it cockroachdb-cluster-0 -- cockroach sql --insecure --execute="INSERT INTO users (id, name, email, password) VALUES ('1', 'name-1', 'email-1', 'pass-1'), ('2', 'name-2', 'email-2', 'pass-2'), ('3', 'name-3', 'email-3', 'pass-3'), ('test', 'testName', 'test@test.test', 'testPass');"

k8s-unit-test: # param: args
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -short -race ${args} ./...  

k8s-integration-test: # param: args
	make k8s-db-seed
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -race ${args} ./...  

k8s-test-gen-coverage:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -coverprofile  cover.out ./...
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go tool cover -html cover.out -o cover.html

k8s-run-dev:
	- $(kubernetes) delete -R -f deployment/k8s/dev/resources/
	$(kubernetes) apply -R -f deployment/k8s/dev/resources/
