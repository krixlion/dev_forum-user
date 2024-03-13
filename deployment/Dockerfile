FROM golang:1.22 AS builder

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

# Must be set in order to avoid dynamic links to libraries unavailable in alpine/scratch images.
ENV CGO_ENABLED=0

RUN mkdir -p /go/src/dev_forum-user
WORKDIR /go/src/dev_forum-user

# Make sure you run `docker build` from the project root or set the context accordingly.
ADD . .

RUN go mod tidy && \
    go mod vendor && \
    go build -o main cmd/main.go

FROM scratch
WORKDIR /app

COPY --from=builder /go/src/dev_forum-user/main /app/main
COPY --from=builder /go/src/dev_forum-user/.env /app/.env

EXPOSE 50051
EXPOSE 2223

CMD [ "/app/main" ]
