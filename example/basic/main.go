package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
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
	server := new(zrpc.Server)
	isClient := flag.Bool("c", false, "if run client")
	flag.Parse()

	if *isClient {
		runRPCClient(server.GetRPCAddress())
		runJSONRPCClient(server.GetJSONRPCAddress())
		return
	}

	arith := new(Arith)

	server.RegisterName("arith", arith)
	if err := server.Listen(); err != nil {
		panic(err)
	}
}

func runRPCClient(address string) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var args interface{}
	args = &Args{7, 8}
	var sum int
	err = client.Call("arith.Sum", args, &sum)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	fmt.Printf("Arith: req -> %v , res -> %v\n", args, sum)
}

func runJSONRPCClient(address string) {
	client, err := jsonrpc.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var args interface{}
	args = &Args{7, 8}
	var sum int
	err = client.Call("arith.Sum", args, &sum)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	fmt.Printf("Arith: req -> %v , res -> %v\n", args, sum)
}

func transferJSONRPCClient(address, method string, params interface{}) (res interface{}, err error) {
	client, dialErr := jsonrpc.Dial("tcp", address)
	if dialErr != nil {
		err = dialErr
		return
	}
	defer client.Close()
	err = client.Call(method, params, &res)
	return
}
