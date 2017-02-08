# nats_to_syslog
Subscribes to NATs message bus and forwards messages to remote Syslog server

# Testing

## Dependencies

```sh
go get github.com/onsi/ginkgo/ginkgo
go get github.com/nats-io/gnatsd

godep go test
```
