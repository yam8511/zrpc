package zrpc

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"net/rpc/jsonrpc"
	"strings"
)

// Input 輸出參數
type Input struct {
	Service string      `json:"service"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
	Address string      `json:"address"`
}

// ErrorDetail 錯誤細節
type ErrorDetail struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Output 輸出參數
type Output struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	ID     int         `json:"id"`
}

// ServeHTTP 服務處理
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if server.debug {
		switch r.URL.EscapedPath() {
		case "/debug/pprof/cmdline":
			pprof.Cmdline(w, r)
			return
		case "/debug/pprof/profile":
			pprof.Profile(w, r)
			return
		case "/debug/pprof/symbol":
			pprof.Symbol(w, r)
			return
		case "/debug/pprof/trace":
			pprof.Trace(w, r)
			return
		}
		if strings.HasPrefix(r.RequestURI, "/debug/") {
			r.URL.Path = strings.Replace(r.URL.Path, "/debug/pprof", "/debug", 1)
			// log.Println("[ZRPC] #1 Req -> ", r.URL.Path)
			r.URL.Path = strings.Replace(r.URL.Path, "/debug", "/debug/pprof", 1)
			// log.Println("[ZRPC] #2 Req -> ", r.URL.Path)
			pprof.Index(w, r)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if r.URL.EscapedPath() == "/services" {
		err := json.NewEncoder(w).Encode(server.Services)
		if err != nil {
			log.Println("[ZRPC] Response Encode Error ->", err)
			return
		}
		return
	}

	var data Input
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Println("[ZRPC] JSON Error ->", err)
		json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    500,
				Message: err.Error(),
				Data:    err,
			},
			ID: data.ID,
		})
		return
	}
	if server.debug {
		log.Println("[ZRPC] Input ->", data)
	}
	address := data.Address
	if address == "" {
		address = server.GetJSONRPCAddress()
	}
	res, err := transferJSONRPCClient(address, data.Method, data.Params)
	if err != nil {
		log.Println("[ZRPC] Call Error ->", err)
		json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    500,
				Message: err.Error(),
				Data:    err,
			},
			ID: data.ID,
		})
		return
	}

	err = json.NewEncoder(w).Encode(Output{
		Result: res,
		Error:  nil,
		ID:     data.ID,
	})
	if err != nil {
		log.Println("[ZRPC] Response Encode Error ->", err)
		return
	}
}

// ServeHTTP 服務處理
func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.EscapedPath() == "/registry" {
		var s []Service
		for _, srv := range proxy.Services {
			s = append(s, srv)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"services": s,
		})
		return
	}
	if proxy.ui && strings.HasPrefix(r.URL.EscapedPath(), "/ui") {
		proxy.WebUI(w, r)
		return
	}

	if r.URL.EscapedPath() != proxy.PrefixPath {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var data Input
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Println("[ZRPC] JSON Error ->", err)
		err = json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    500,
				Message: err.Error(),
				Data:    err,
			},
			ID: data.ID,
		})
		if err != nil {
			log.Println("[ZRPC] Response Encode Error ->", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if proxy.debug {
		log.Printf("[ZRPC] Server (%s), Redirect to %s", data.Service, data.Address)
	}
	service, ok := proxy.Services[data.Service]
	if !ok {
		err = json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    500,
				Message: "Service Not Found",
				Data:    "Service: " + data.Service,
			},
			ID: data.ID,
		})
		if err != nil {
			log.Println("[ZRPC] Response Encode Error ->", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	res, err := transferJSONRPCClient(service.RPCAddress, data.Method, data.Params)
	if err != nil {
		log.Println("[ZRPC] Call Error ->", err)
		err = json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    500,
				Message: err.Error(),
				Data:    err,
			},
			ID: data.ID,
		})
		if err != nil {
			log.Println("[ZRPC] Response Encode Error ->", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = json.NewEncoder(w).Encode(Output{
		Result: res,
		Error:  nil,
		ID:     data.ID,
	})
	if err != nil {
		log.Println("[ZRPC] Response Encode Error ->", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
