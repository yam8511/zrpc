package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Server 伺服端
type Server struct {
	RPCAddr     string
	RPCNet      net.Listener
	JSONRPCAddr string
	JSONRPCNet  net.Listener
	HTTPAddr    string
	HTTPNet     net.Listener
	HTTPServer  *http.Server
	Timeout     int64
}

// Init 初始化
func (server *Server) Init() error {
	if server.RPCNet == nil || server.RPCNet.Addr().Network() == "" {
		l, e := net.Listen("tcp", server.GetRPCAddress())
		if e != nil {
			return e
		}
		server.RPCNet = l
	}

	if server.JSONRPCNet == nil || server.JSONRPCNet.Addr().Network() == "" {
		l, e := net.Listen("tcp", server.GetJSONRPCAddress())
		if e != nil {
			return e
		}
		server.JSONRPCNet = l
	}

	if server.HTTPNet == nil || server.HTTPNet.Addr().Network() == "" {
		l, e := net.Listen("tcp", server.GetHTTPAddress())
		if e != nil {
			return e
		}
		server.HTTPNet = l
	}

	if server.HTTPServer == nil {
		server.HTTPServer = &http.Server{
			Handler: new(Handler),
		}
	}

	return nil
}

// SetRPCNet 設定RPC網路
func (server *Server) SetRPCNet(n net.Listener) {
	server.RPCNet = n
}

// SetRPCAddress 設定RPC連線網址
func (server *Server) SetRPCAddress(addr string) {
	server.RPCAddr = addr
}

// SetJSONRPCNet 設定JSONRPC網路
func (server *Server) SetJSONRPCNet(n net.Listener) {
	server.JSONRPCNet = n
}

// SetJSONRPCAddress 設定JSONRPC連線網址
func (server *Server) SetJSONRPCAddress(addr string) {
	server.JSONRPCAddr = addr
}

// SetHTTPNet 設定HTTP網路
func (server *Server) SetHTTPNet(n net.Listener) {
	server.HTTPNet = n
}

// SetHTTPAddress 設定HTTP連線網址
func (server *Server) SetHTTPAddress(addr string) {
	server.HTTPAddr = addr
}

// SetHTTPServer 設定HTTP-Server
func (server *Server) SetHTTPServer(h *http.Server) {
	server.HTTPServer = h
}

// GetRPCAddress 取RPC的連線網址
func (server *Server) GetRPCAddress() string {
	if server.RPCAddr == "" {
		return ":50051"
	}
	return server.RPCAddr
}

// GetJSONRPCAddress 取JSONRPC的連線網址
func (server *Server) GetJSONRPCAddress() string {
	if server.JSONRPCAddr == "" {
		return ":50052"
	}
	return server.JSONRPCAddr
}

// GetHTTPAddress 取HTTP的連線網址
func (server *Server) GetHTTPAddress() string {
	if server.HTTPAddr == "" {
		return ":8000"
	}
	return server.HTTPAddr
}

// SetTimeout 設定連線逾時秒數
func (server *Server) SetTimeout(second int64) {
	server.Timeout = second
}

// Register 註冊服務
func (server *Server) Register(service interface{}) {
	rpc.Register(service)
}

// RegisterName 註冊服務
func (server *Server) RegisterName(name string, service interface{}) {
	rpc.RegisterName(name, service)
}

// Listen 監聽連線
func (server *Server) Listen() error {
	// 檢查連線設定
	if server.RPCNet == nil || server.RPCNet.Addr().Network() == "" {
		l, e := net.Listen("tcp", server.GetRPCAddress())
		if e != nil {
			return e
		}
		server.RPCNet = l
	}

	if server.JSONRPCNet == nil || server.JSONRPCNet.Addr().Network() == "" {
		l, e := net.Listen("tcp", server.GetJSONRPCAddress())
		if e != nil {
			return e
		}
		server.JSONRPCNet = l
	}

	if server.HTTPNet == nil || server.HTTPNet.Addr().Network() == "" {
		l, e := net.Listen("tcp", server.GetHTTPAddress())
		if e != nil {
			return e
		}
		server.HTTPNet = l
	}

	if server.HTTPServer == nil {
		server.HTTPServer = &http.Server{
			Handler: new(Handler),
		}
	}

	// 設置關閉機制
	var (
		err error
		sig = make(chan os.Signal)
		c   = make(chan int)
		e   = make(chan error)
		wg  = new(sync.WaitGroup)
	)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		select {
		case s := <-sig:
			log.Printf("... Receive signal, shutdown by ... %v", s)
			close(c)
			server.RPCNet.Close()
			server.JSONRPCNet.Close()
			server.HTTPServer.Close()
		case err = <-e:
			log.Printf("... Listen get error ... %s", err.Error())
			close(c)
			server.RPCNet.Close()
			server.JSONRPCNet.Close()
			server.HTTPServer.Close()
		}
	}()
	wg.Add(3)

	// RPC
	go func() {
		defer wg.Done()
		for {
			conn, err := server.RPCNet.Accept()
			if err != nil {
				select {
				case <-c:
					return
				default:
					log.Println("Error: accept rpc connection ->", err)
					e <- err
				}
				continue
			}
			log.Println("Accept rpc connection")
			// 設定連線timeout
			if server.Timeout > 0 {
				conn.SetDeadline(time.Now().Add(time.Second * time.Duration(server.Timeout)))
			}
			go rpc.ServeConn(conn)
		}
	}()

	// JSON-RPC
	go func() {
		defer wg.Done()
		for {
			conn, err := server.JSONRPCNet.Accept()
			if err != nil {
				select {
				case <-c:
					return
				default:
					log.Println("Error: accept jsonrpc connection ->", err)
					e <- err
				}
				continue
			}
			log.Println("Accept jsonrpc connection")
			// 設定連線timeout
			if server.Timeout > 0 {
				conn.SetDeadline(time.Now().Add(time.Second * time.Duration(server.Timeout)))
			}
			go jsonrpc.ServeConn(conn)
		}
	}()

	// HTTP
	go func() {
		defer wg.Done()
		err := server.HTTPServer.Serve(server.HTTPNet)
		if err != nil {
			select {
			case <-c:
				return
			default:
				log.Println("Error: accept http connection ->", err)
				e <- err
			}
		}
	}()

	log.Println("RPC Server Listening ... ", server.RPCNet.Addr().Network(), server.RPCNet.Addr().String())
	log.Println("JSON-RPC Server Listening ... ", server.JSONRPCNet.Addr().Network(), server.JSONRPCNet.Addr().String())
	log.Println("HTTP Server Listening ... ", server.HTTPNet.Addr().Network(), server.HTTPNet.Addr().String())

	wg.Wait()
	return err
}
