# golms

## Prerequisites
Make sure to have the following:
- go
- LLM models downloaded in a directory ~/golms/mlx_lm OR golms/ollama
- Model Servers downloaded

### Default: Uses cache, ignores vendor/
```go
go build
```
Looks in ~/go/pkg/mod/

### Vendor mode: Uses vendor/
```go
go build -mod=vendor
```
Looks in ./vendor/