# Simple Test of Memory allocation

```bash
cd main
go run .
Ctrl+C
go build
go tool pprof -http 127.0.0.1:8080 ./main.exe mem.prof
```
or
run with available binary:
```bash
cd main
go tool pprof -http 127.0.0.1:8080 ./main mem.prof
```

Then open `localhost:8080` for ui