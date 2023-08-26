FROM golang:1.21 AS build-stage
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/producer

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /


COPY --from=build-stage /app/producer /app/producer

USER nonroot:nonroot
ENTRYPOINT [ "/app/producer" ]