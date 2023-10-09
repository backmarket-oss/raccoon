FROM golang:1.21.2-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY internal ./internal
COPY cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /raccoon

FROM gcr.io/distroless/static-debian11
COPY --from=builder /raccoon /raccoon
ENTRYPOINT ["./raccoon"]
