package zrpc

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Proxy 代理伺服
type Proxy struct {
	PrefixPath string
	Services   map[string]Service
	HTTPAddr   string
	HTTPNet    net.Listener
	HTTPServer *http.Server
	Timeout    int64
	UI         bool
	mx         *sync.RWMutex
}

// NewProxy 建立一個伺服器
func NewProxy() (p *Proxy) {
	p = &Proxy{
		Services: map[string]Service{},
		mx:       new(sync.RWMutex),
	}
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
			Handler: proxy,
		}
	}
	return nil
}

// AddService 新增服務
func (proxy *Proxy) AddService(name, addr string) {
	proxy.mx.Lock()
	defer proxy.mx.Unlock()
	service, ok := proxy.Services[name]
	if ok {
		service.Address = addr
	} else {
		service = Service{
			Name:    name,
			Address: addr,
		}
	}
	proxy.Services[name] = service
}

// SetHTTPNet 設定HTTP網路
func (proxy *Proxy) SetHTTPNet(n net.Listener) {
	proxy.HTTPNet = n
}

// SetHTTPAddress 設定HTTP連線網址
func (proxy *Proxy) SetHTTPAddress(addr string) {
	proxy.HTTPAddr = addr
}

// SetHTTPServer 設定HTTP-Server
func (proxy *Proxy) SetHTTPServer(h *http.Server) {
	proxy.HTTPServer = h
}

// GetHTTPAddress 取HTTP的連線網址
func (proxy *Proxy) GetHTTPAddress() string {
	if proxy.HTTPAddr == "" {
		return ":8081"
	}
	return proxy.HTTPAddr
}

// SetTimeout 設定連線逾時秒數
func (proxy *Proxy) SetTimeout(second int64) {
	proxy.Timeout = second
}

// EnableWebUI 啟動界面
func (proxy *Proxy) EnableWebUI(enable bool) {
	proxy.UI = enable
}

// SetPrefixPath 設定前綴
func (proxy *Proxy) SetPrefixPath(path string) {
	proxy.PrefixPath = path
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
		e   = make(chan error)
		wg  = new(sync.WaitGroup)
	)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		select {
		case s := <-sig:
			log.Printf("... Receive signal, shutdown by ... %v", s)
			close(c)
			proxy.HTTPServer.Close()
		case err = <-e:
			log.Printf("... Listen get error ... %s", err.Error())
			close(c)
			proxy.HTTPServer.Close()
		}
	}()
	wg.Add(1)

	// HTTP
	go func() {
		defer wg.Done()
		err := proxy.HTTPServer.Serve(proxy.HTTPNet)
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
	log.Println("HTTP Server Listening ... ", proxy.HTTPNet.Addr().Network(), proxy.HTTPNet.Addr().String())

	wg.Wait()
	return err
}
