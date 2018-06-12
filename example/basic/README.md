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
2018/06/13 01:24:27 RPC Server Listening ...  tcp [::]:50051
2018/06/13 01:24:27 JSON-RPC Server Listening ...  tcp [::]:50052
2018/06/13 01:24:27 HTTP Server Listening ...  tcp [::]:8000
2018/06/13 01:24:33 Accept rpc connection
2018/06/13 01:24:34 Accept jsonrpc connection
```

2. Open Another Terminal
```shell
$ ./app -c
Arith: req -> &{7 8} , res -> 15
Arith: req -> &{7 8} , res -> 15
```