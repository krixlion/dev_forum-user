# dev_forum-user
User-service registers user identities in the dev_forum system.

It's dependent on:
  - [CockroachDB](https://www.cockroachlabs.com/docs/cockroachcloud/quickstart) for persistent user storage.
  - [RabbitMQ](https://www.rabbitmq.com/#getstarted) for asynchronous communication with the other components in the domain.
  - [OtelCollector](https://opentelemetry.io/docs/collector) for receiving and forwarding telemetry data.

## Set up
Rename `.env.example` to `.env` and fill in missing values.

### Using Go command
You need a working [Go environment](https://go.dev/doc/install).
```shell
go mod tidy
go mod vendor
go build cmd/main.go
```

### On Docker
You need a working [Docker environment](https://docs.docker.com/engine).

You can use the Dockerfile located in `deployment/` to build and run the service on a docker container.

```shell
make build-image version=latest` 
``` 

```shell
docker run -p 50051:50051 -p 2223:2223 krixlion/dev_forum-user:0.1.0
```
### On Kubernetes (recommended)
You need a working [Kubernetes environment](https://kubernetes.io/docs/setup) with [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization).

Kubernetes resources are defined in `deployment/k8s` and deployed using [Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).
Currently there are `stage` and `dev` overlays available and include any needed resources and configs.

Use `make` to apply manifests for dev_forum-user and needed DBs for either dev or stage environment.
```shell
make k8s-run overlay=<dev/stage>
```
```shell
# To delete
make k8s-stop overlay=dev
```

## Testing

Run unit and integration tests using Go command.
```shell
#  Add `-short` flag to skip integration tests.
go test ./... -race
```

Generate coverage report using `go tool cover`.
```
go test -coverprofile  cover.out ./...
go tool cover -html cover.out -o cover.html
```

If the service is deployed on kubernetes you can use `make`.
```shell
make k8s-integration-test
```
or
```shell
make k8s-unit-test
```

## Documentation 
For in-detail documentation refer to the [Wiki](https://github.com/krixlion/dev_forum-user/wiki).

## API
Service is exposing [gRPC](https://grpc.io/docs/what-is-grpc/introduction) API.

Regenerate `pb` packages after making changes to any of the `.proto` files located in `api/`.
You can use [go-grpc-gen](https://github.com/krixlion/go-grpc-gen) containerized tool with `make grpc-gen`.
