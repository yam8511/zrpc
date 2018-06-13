# Basic Example

1. Open Terminal
```shell
$ go build -o app
$ ./app
=============================
註冊新服務 -> arith
方法 Diff -> func(*main.Args, *int) error
方法 Sum -> func(*main.Args, *int) error
=============================
2018/06/14 01:27:08 HTTP Server Listening ...  tcp [::]:8081
2018/06/14 01:27:08 RPC Server Listening ...  tcp [::]:50051
2018/06/14 01:27:08 JSON-RPC Server Listening ...  tcp [::]:50052
2018/06/14 01:27:08 HTTP Server Listening ...  tcp [::]:8000
```

2. Open Another Terminal
```shell
$ curl -X POST \
  http://127.0.0.1:8081/rpc \
  -H 'Content-Type: application/json' \
  -d '{
	"id": 1,
	"service": "Arith",
	"method":"arith.Sum",
	"params": {
		"A": 1,
		"B": 2
	}
  }'
{"result":3,"error":null,"id":1}
```