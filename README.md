# riemann-http

### Compile and execute

- To compile:

```bash
go build -ldflags="-w -s" -o server cmd/main.go
```

- To execute:

```bash
AUTH_USER=user AUTH_PASSWORD=password ./server
```