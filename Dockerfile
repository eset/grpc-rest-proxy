FROM golang:1.22-alpine AS build

RUN apk add --no-cache --no-progress git make

WORKDIR /app
COPY . .
RUN make build-only

FROM alpine:3 AS runtime

WORKDIR /app
COPY --from=build /app/grpc-rest-proxy .

EXPOSE 8080

ENTRYPOINT ["/app/grpc-rest-proxy"]
