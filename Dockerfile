FROM golang:1.14 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY pkg/ pkg/
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o k8s-node-labeler main.go

FROM alpine:3.10
COPY --from=builder /workspace/k8s-node-labeler /
RUN chmod +x /k8s-node-labeler
RUN adduser -DH runner
USER runner
ENTRYPOINT ["./k8s-node-labeler"]