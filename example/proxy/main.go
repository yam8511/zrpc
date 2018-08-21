package main

import (
	"time"

	"github.com/yam8511/zrpc"
)

// Arith 數學運算
type Arith int

// Args 參數
type Args struct {
	A, B int
}

// Sum 總和
func (t *Arith) Sum(args *Args, sum *int) error {
	if args.A == 0 && args.B == 0 {
		return zrpc.NewZrpcError("422", "缺少參數", map[string]int{
			"A": args.A,
			"B": args.B,
		})
	}
	time.Sleep(time.Second * 1)
	*sum = args.A + args.B
	return nil
}

// Diff 差和
func (t *Arith) Diff(args *Args, diff *int) error {
	if args.A == 0 && args.B == 0 {
		return zrpc.NewZrpcError("422", "缺少參數", map[string]int{
			"A": args.A,
			"B": args.B,
		})
	}
	*diff = args.A - args.B
	return nil
}

func main() {
	arith := new(Arith)
	server := zrpc.NewServer()
	server.RegisterName("arith", arith)
	go func() {
		if err := server.Listen(); err != nil {
			panic(err)
		}
	}()

	proxy := zrpc.NewProxy()
	proxy.AddService("Arith", server.GetJSONRPCAddress(), server.GetHTTPAddress())
	err := proxy.Listen()
	if err != nil {
		panic(err)
	}
}
