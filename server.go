package zrpc

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"strconv"
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
	Services    []Service
	kind        string
	timeout     int64
	debug       bool
}

// NewServer 建立一個伺服器
func NewServer() *Server {
	var (
		server   *Server
		rpcKind  = os.Getenv("ZRPC_SERVER")
		rpcAddr  = os.Getenv("ZRPC_SERVER_ADDRESS")
		httpAddr = os.Getenv("ZRPC_HTTP_ADDRESS")
	)

	if rpcKind == "rpc" {
		server = &Server{
			kind:     rpcKind,
			RPCAddr:  rpcAddr,
			HTTPAddr: httpAddr,
		}
	} else {
		server = &Server{
			kind:        "jsonrpc",
			JSONRPCAddr: rpcAddr,
			HTTPAddr:    httpAddr,
		}
	}

	// 檢查Timeout環境變數
	if st := os.Getenv("ZRPC_TIMEOUT"); st != "" {
		t, err := strconv.Atoi(st)
		if err != nil {
			server.SetTimeout(0)
		} else {
			server.SetTimeout(int64(t))
		}
	}

	// 檢查除錯模式
	server.DebugMode(os.Getenv("ZRPC_DEBUG_MODE") == "true")
	return server
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
			Handler: server,
		}
	}

	return nil
}

// DebugMode 設定Debug模式
func (server *Server) DebugMode(debug bool) *Server {
	if debug {
		log.Println("[ZRPC] Server Debug Mode: On")
	} else {
		log.Println("[ZRPC] Server Debug Mode: Off")
	}
	server.debug = debug
	return server
}

// SetServer 設定伺服器
func (server *Server) SetServer(s string) *Server {
	if s == "rpc" || s == "jsonrpc" {
		server.kind = s
	} else {
		panic("server is wrong")
	}
	return server
}

// SetRPCNet 設定RPC網路
func (server *Server) SetRPCNet(n net.Listener) *Server {
	server.RPCNet = n
	return server
}

// SetRPCAddress 設定RPC連線網址
func (server *Server) SetRPCAddress(addr string) *Server {
	server.RPCAddr = addr
	return server
}

// SetJSONRPCNet 設定JSONRPC網路
func (server *Server) SetJSONRPCNet(n net.Listener) *Server {
	server.JSONRPCNet = n
	return server
}

// SetJSONRPCAddress 設定JSONRPC連線網址
func (server *Server) SetJSONRPCAddress(addr string) *Server {
	server.JSONRPCAddr = addr
	return server
}

// SetHTTPNet 設定HTTP網路
func (server *Server) SetHTTPNet(n net.Listener) *Server {
	server.HTTPNet = n
	return server
}

// SetHTTPAddress 設定HTTP連線網址
func (server *Server) SetHTTPAddress(addr string) *Server {
	server.HTTPAddr = addr
	return server
}

// SetHTTPServer 設定HTTP-Server
func (server *Server) SetHTTPServer(h *http.Server) *Server {
	server.HTTPServer = h
	return server
}

// GetRPCAddress 取RPC的連線網址
func (server *Server) GetRPCAddress() string {
	if server.RPCAddr == "" {
		if addr := os.Getenv("ZRPC_SERVER_ADDRESS"); addr != "" {
			return addr
		}
		return ":50051"
	}
	return server.RPCAddr
}

// GetJSONRPCAddress 取JSONRPC的連線網址
func (server *Server) GetJSONRPCAddress() string {
	if server.JSONRPCAddr == "" {
		if addr := os.Getenv("ZRPC_SERVER_ADDRESS"); addr != "" {
			return addr
		}
		return ":50052"
	}
	return server.JSONRPCAddr
}

// GetHTTPAddress 取HTTP的連線網址
func (server *Server) GetHTTPAddress() string {
	if server.HTTPAddr == "" {
		if addr := os.Getenv("ZRPC_HTTP_ADDRESS"); addr != "" {
			return addr
		}
		return ":8000"
	}
	return server.HTTPAddr
}

// SetTimeout 設定連線逾時秒數
func (server *Server) SetTimeout(second int64) *Server {
	server.timeout = second
	return server
}

// Register 註冊服務
func (server *Server) Register(service interface{}) *Server {
	rpc.Register(service)
	name, methods := ReflectMethod(service)
	server.Services = append(server.Services, Service{
		Name:    name,
		Methods: methods,
	})
	if server.debug {
		log.Println("[ZRPC] =============================")
		log.Println("[ZRPC] 註冊新服務 ->", name)
		for methodName, methodType := range methods {
			log.Printf("[ZRPC] 方法 %s -> %s\n", methodName, methodType)
		}
		log.Println("[ZRPC] =============================")
	}
	return server
}

// RegisterName 註冊服務
func (server *Server) RegisterName(name string, service interface{}) *Server {
	rpc.RegisterName(name, service)
	_, methods := ReflectMethod(service)
	server.Services = append(server.Services, Service{
		Name:    name,
		Methods: methods,
	})
	if server.debug {
		log.Println("[ZRPC] =============================")
		log.Println("[ZRPC] 註冊新服務 ->", name)
		for methodName, methodType := range methods {
			log.Printf("[ZRPC] 方法 %s -> %s\n", methodName, methodType)
		}
		log.Println("[ZRPC] =============================")
	}
	return server
}

// Listen 監聽連線
func (server *Server) Listen() error {
	// 檢查連線設定
	if err := server.Init(); err != nil {
		return err
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
			log.Printf("[ZRPC] ... Receive signal, shutdown by ... %v", s)
			close(c)
			server.RPCNet.Close()
			server.JSONRPCNet.Close()
			server.HTTPServer.Close()
		case err = <-e:
			log.Printf("[ZRPC] ... Listen get error ... %s", err.Error())
			close(c)
			server.RPCNet.Close()
			server.JSONRPCNet.Close()
			server.HTTPServer.Close()
		}
	}()

	switch server.kind {
	case "rpc":
		// RPC
		wg.Add(1)
		log.Println("[ZRPC] RPC Server Listening ... ", server.RPCNet.Addr().Network(), server.RPCNet.Addr().String())
		go func() {
			defer wg.Done()
			for {
				conn, err := server.RPCNet.Accept()
				if err != nil {
					select {
					case <-c:
						return
					default:
						log.Println("[ZRPC] Error: accept rpc connection ->", err)
						e <- err
					}
					continue
				}
				log.Println("[ZRPC] Accept rpc connection")
				// 設定連線timeout
				if server.timeout > 0 {
					conn.SetDeadline(time.Now().Add(time.Second * time.Duration(server.timeout)))
				}
				go rpc.ServeConn(conn)
			}
		}()
		break
	case "jsonrpc":
		// JSON-RPC
		wg.Add(1)
		log.Println("[ZRPC] JSON-RPC Server Listening ... ", server.JSONRPCNet.Addr().Network(), server.JSONRPCNet.Addr().String())
		go func() {
			defer wg.Done()
			for {
				conn, err := server.JSONRPCNet.Accept()
				if err != nil {
					select {
					case <-c:
						return
					default:
						log.Println("[ZRPC] Error: accept jsonrpc connection ->", err)
						e <- err
					}
					continue
				}
				log.Println("[ZRPC] Accept jsonrpc connection")
				// 設定連線timeout
				if server.timeout > 0 {
					conn.SetDeadline(time.Now().Add(time.Second * time.Duration(server.timeout)))
				}
				go jsonrpc.ServeConn(conn)
			}
		}()
		break
	}

	// HTTP
	wg.Add(1)
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
	log.Println("[ZRPC] HTTP Server Listening ... ", server.HTTPNet.Addr().Network(), server.HTTPNet.Addr().String())

	wg.Wait()
	return err
}
