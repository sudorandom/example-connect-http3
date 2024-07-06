# Example of gRPC+HTTP/3

## TLS cert
This is the command used to create the self-signed cert:

```shell
openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
    -subj "/C=DK/L=Copenhagen/O=kmcd/CN=local.kmcd.dev" \
    -keyout cert.key  -out cert.crt
```

## Starting the server
```shell
go run server/main.go
```

## Running the client
```shell
go run client/main.go
```
