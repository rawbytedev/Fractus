# Simple Test of Memory allocation

```bash
cd cmd/fractus-server
go run .
go build -o fractus
Ctrl+C
go tool pprof -http:=8080 ./fractus mem.prof
```
or
run with available binary:
```bash
cd cmd/fractus-server
go tool pprof -http:=8080 ./fractus mem.prof
```

Then open `localhost:8080` for ui