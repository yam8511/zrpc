package zrpc

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Proxy 代理伺服
type Proxy struct {
	PrefixPath string
	Services   map[string]Service
	HTTPAddr   string
	HTTPNet    net.Listener
	HTTPServer *http.Server
	timeout    int64
	ui         bool
	debug      bool
	mx         *sync.RWMutex
}

// NewProxy 建立一個伺服器
func NewProxy() (p *Proxy) {
	p = &Proxy{
		Services: map[string]Service{},
		mx:       new(sync.RWMutex),
	}
	p.SetHTTPAddress(os.Getenv("ZRPC_PROXY_ADDRESS"))
	p.EnableWebUI(os.Getenv("ZRPC_ENABLE_UI") == "true")
	p.DebugMode(os.Getenv("ZRPC_DEBUG_MODE") == "true")
	return
}

// Init 初始化
func (proxy *Proxy) Init() error {
	if proxy.PrefixPath == "" {
		proxy.PrefixPath = "/rpc"
	}
	if proxy.HTTPNet == nil || proxy.HTTPNet.Addr().Network() == "" {
		l, e := net.Listen("tcp", proxy.GetHTTPAddress())
		if e != nil {
			return e
		}
		proxy.HTTPNet = l
	}

	if proxy.HTTPServer == nil {
		proxy.HTTPServer = &http.Server{
			Handler:      proxy,
			WriteTimeout: time.Duration(proxy.timeout),
			ReadTimeout:  time.Duration(proxy.timeout),
		}
	}
	return nil
}

// AddService 新增服務
func (proxy *Proxy) AddService(name, rpcAddr, httpAddr string) *Proxy {
	if proxy.debug {
		log.Println("[ZRPC] =============================")
		log.Println("[ZRPC] 註冊新服務 ->", name)
		log.Println("[ZRPC] TCP 服務位址 ->", rpcAddr)
		log.Println("[ZRPC] HTTP 服務位址 ->", httpAddr)
		log.Println("[ZRPC] =============================")
	}
	proxy.mx.Lock()
	defer proxy.mx.Unlock()
	service, ok := proxy.Services[name]
	if ok {
		service.RPCAddress = rpcAddr
		service.HTTPAddress = httpAddr
	} else {
		service = Service{
			Name:        name,
			RPCAddress:  rpcAddr,
			HTTPAddress: httpAddr,
		}
	}
	proxy.Services[name] = service
	return proxy
}

// DebugMode 設定Debug模式
func (proxy *Proxy) DebugMode(debug bool) *Proxy {
	if debug {
		log.Println("[ZRPC] Proxy Debug Mode: On")
	} else {
		log.Println("[ZRPC] Proxy Debug Mode: Off")
	}
	proxy.debug = debug
	return proxy
}

// SetHTTPNet 設定HTTP網路
func (proxy *Proxy) SetHTTPNet(n net.Listener) *Proxy {
	proxy.HTTPNet = n
	return proxy
}

// SetHTTPAddress 設定HTTP連線網址
func (proxy *Proxy) SetHTTPAddress(addr string) *Proxy {
	proxy.HTTPAddr = addr
	return proxy
}

// SetHTTPServer 設定HTTP-Server
func (proxy *Proxy) SetHTTPServer(h *http.Server) *Proxy {
	proxy.HTTPServer = h
	return proxy
}

// GetHTTPAddress 取HTTP的連線網址
func (proxy *Proxy) GetHTTPAddress() string {
	if proxy.HTTPAddr == "" {
		if addr := os.Getenv("ZRPC_PROXY_ADDRESS"); addr != "" {
			return addr
		}
		return ":8081"
	}
	return proxy.HTTPAddr
}

// SetTimeout 設定連線逾時秒數
func (proxy *Proxy) SetTimeout(second int64) *Proxy {
	proxy.timeout = second
	return proxy
}

// EnableWebUI 啟動界面
func (proxy *Proxy) EnableWebUI(enable bool) *Proxy {
	if enable {
		log.Println("[ZRPC] Web UI: On")
	} else {
		log.Println("[ZRPC] Web UI: Off")
	}
	proxy.ui = enable
	return proxy
}

// SetPrefixPath 設定前綴
func (proxy *Proxy) SetPrefixPath(path string) *Proxy {
	proxy.PrefixPath = path
	return proxy
}

// Listen 監聽服務
func (proxy *Proxy) Listen() error {
	// 檢查連線設定
	if err := proxy.Init(); err != nil {
		return err
	}

	// 設置關閉機制
	var (
		err error
		sig = make(chan os.Signal)
		c   = make(chan int)
	)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		s := <-sig
		log.Printf("[ZRPC] ... Receive signal, shutdown by ... %v", s)
		close(c)
		proxy.HTTPServer.Close()
	}()

	log.Println("[ZRPC] HTTP Server Listening ... ", proxy.HTTPNet.Addr().Network(), proxy.HTTPNet.Addr().String())
	err = proxy.HTTPServer.Serve(proxy.HTTPNet)
	if err != nil {
		select {
		case <-c:
			return nil
		default:
			log.Printf("[ZRPC] ... Listen get error ... %s", err.Error())
			return err
		}
	}
	return nil
}
