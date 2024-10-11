# Project gocache-block-ips

This is an project for a technical challenge. Check how to Run the application and test it below.

## Getting Started

First you will have to execute a `curl` after start the application by running the following command in two terminals:

1. `make docker-run`
2. `make watch`

If those commands you will be able to run the start the application itself, now to block some IPs you can execute the following curl request:

```bash
curl -X POST http://localhost:8080/blockips \
  -H "Content-Type: application/json" \
  -d '{"ips": ["192.168.1.1", "10.0.0.2"]}'
```

This command will block those IPs, now to test if the request will be blocked or not you can execute the shell script in the root of the project by running and passing the ip like in the one below for success:

`make test-ip-success`

and another one below for error:

`make test-ip-error`

Or you can set do it manually by running the following commands:

1. `chmod +x ./make-post-ip.sh`
2. `./make-post-ip.sh 192.168.1.1`

and depending on you request the response you be different.

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run ip test success:
```bash
make test-ip-success
```
Run ip test error:
```bash
make test-ip-error
```


Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
