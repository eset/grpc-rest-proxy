FROM golang:1.23-alpine AS build

RUN apk add --no-cache --no-progress git make

WORKDIR /app
COPY . .
RUN make build-example-grpc-server

FROM alpine:3 AS runtime

WORKDIR /app
COPY --from=build /app/example-grpc-server .

EXPOSE 50051

ENTRYPOINT ["/app/example-grpc-server"]
