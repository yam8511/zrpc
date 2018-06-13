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
	time.Sleep(time.Second * 1)
	*sum = args.A + args.B
	return nil
}

// Diff 差和
func (t *Arith) Diff(args *Args, diff *int) error {
	*diff = args.A - args.B
	return nil
}

func main() {
	arith := new(Arith)
	server := new(zrpc.Server)
	server.RegisterName("arith", arith)
	go func() {
		if err := server.Listen(); err != nil {
			panic(err)
		}
	}()

	proxy := zrpc.NewProxy()
	proxy.AddService("Arith", server.GetJSONRPCAddress())
	err := proxy.Listen()
	if err != nil {
		panic(err)
	}
}
