# Example of gRPC+HTTP/3

## TLS cert
This is the command used to create the self-signed cert:

```shell
openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
    -subj "/C=DK/L=Copenhagen/O=kmcd/CN=local.kmcd.dev" \
    -keyout cert.key  -out cert.crt
```

## Starting the server (HTTP3 only)
```shell
go run server-single/main.go
```

## Starting the server (HTTP/1.1, HTTP/2 and HTTP/3)
```shell
go run server-multi/main.go
```

## Running the client (http stdlib)
```shell
go run client-http/main.go
```

## Running the client (connect)
```shell
go run client-connect/main.go
```
