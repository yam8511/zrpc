# Basic Example

1. Open Terminal
```shell
$ go build -o app
$ ZRPC_DEBUG_MODE=true ./app
2018/06/16 14:25:54 [ZRPC]Server Debug Mode: On
2018/06/16 14:25:54 [ZRPC] =============================
2018/06/16 14:25:54 [ZRPC] 註冊新服務 -> arith
2018/06/16 14:25:54 [ZRPC] 方法 Diff -> func(*main.Args, *int) error
2018/06/16 14:25:54 [ZRPC] 方法 Sum -> func(*main.Args, *int) error
2018/06/16 14:25:54 [ZRPC] =============================
2018/06/16 14:25:54 [ZRPC] JSON-RPC Server Listening ...  tcp [::]:50052
2018/06/16 14:25:54 [ZRPC] HTTP Server Listening ...  tcp [::]:8000
2018/06/16 14:26:58 [ZRPC] Accept jsonrpc connection
```

2. Open Another Terminal
```shell
$ ./app -c
2018/06/16 14:26:58 [ZRPC]Server Debug Mode: Off
Arith: req -> &{7 8} , res -> 15
```