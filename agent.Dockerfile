FROM golang:latest AS build_step
LABEL stage=builder
ENV GO111MODULE=on
WORKDIR  /go/src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/build/main /go/src/cmd/main.go

FROM docker:latest
WORKDIR /app
COPY --from=build_step /go/build/main /app/main
RUN chmod +x /app/main
EXPOSE 3000/tcp
ENTRYPOINT /app/main
