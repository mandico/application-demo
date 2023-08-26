FROM golang:1.21 AS build
WORKDIR /app
COPY . .

#RUN go mod download

RUN go build -o /app/consumer main.go 

FROM alpine:latest
WORKDIR /app

COPY --from=build /app /app

USER nonroot:nonroot
ENTRYPOINT [ "/app/consumer" ]