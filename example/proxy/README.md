# Proxy Example

1. Open Terminal
```shell
$ go build -o app
$ ZRPC_ENABLE_UI=true ZRPC_DEBUG_MODE=true ./app
2018/06/16 14:40:14 [ZRPC] Server Debug Mode: On
2018/06/16 14:40:14 [ZRPC] =============================
2018/06/16 14:40:14 [ZRPC] 註冊新服務 -> arith
2018/06/16 14:40:14 [ZRPC] 方法 Sum -> func(*main.Args, *int) error
2018/06/16 14:40:14 [ZRPC] 方法 Diff -> func(*main.Args, *int) error
2018/06/16 14:40:14 [ZRPC] =============================
2018/06/16 14:40:14 [ZRPC] Web UI: On
2018/06/16 14:40:14 [ZRPC] Proxy Debug Mode: On
2018/06/16 14:40:14 [ZRPC] =============================
2018/06/16 14:40:14 [ZRPC] 註冊新服務 -> Arith
2018/06/16 14:40:14 [ZRPC] TCP 服務位址 -> :50052
2018/06/16 14:40:14 [ZRPC] HTTP 服務位址 -> :8000
2018/06/16 14:40:14 [ZRPC] =============================
2018/06/16 14:40:14 [ZRPC] HTTP Server Listening ...  tcp [::]:8081
2018/06/16 14:40:14 [ZRPC] JSON-RPC Server Listening ...  tcp [::]:50052
2018/06/16 14:40:14 [ZRPC] HTTP Server Listening ...  tcp [::]:8000
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

3. Open the browser, see http://127.0.0.1:8081/ui