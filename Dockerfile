FROM golang:1.25.3-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o proxy

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/proxy /proxy
EXPOSE 8000
ENTRYPOINT ["/proxy"]
