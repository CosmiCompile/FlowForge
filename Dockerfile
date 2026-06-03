FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 go build -o /out/flowforge-server ./cmd/flowforge-server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/flowforge-server /usr/local/bin/flowforge-server
EXPOSE 8080
ENTRYPOINT ["flowforge-server"]
