package zrpc

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"net/rpc/jsonrpc"
	"strings"
)

// ServeHTTP 服務處理
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		w.WriteHeader(http.StatusOK)
		return
	}

	ip := r.RemoteAddr
	server.httpIn <- ip
	defer func(ip string) {
		server.httpOut <- ip
	}(ip)

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
			log.Println("[ZRPC] Response Error ->", err)
			return
		}
		return
	}

	var data Input
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		err = json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    "500",
				Message: err.Error(),
				Data:    err,
			},
			ID: data.ID,
		})
		if err != nil {
			log.Println("[ZRPC] Response Error ->", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	address := data.Address
	if address == "" {
		address = server.GetJSONRPCAddress()
	}
	res, err := transferJSONRPCClient(address, data.Method, data.Params)
	if err != nil {
		output := Output{
			Result: nil,
			Error:  nil,
			ID:     data.ID,
		}

		jsonrpcErr, yes := IsZrpcError(err)
		if yes {
			output.Error = jsonrpcErr
		} else {
			log.Println("[ZRPC] JSON DECODE Error ->", err)
			output.Error = ErrorDetail{
				Code:    "500",
				Message: err.Error(),
				Data:    err,
			}
		}

		err = json.NewEncoder(w).Encode(output)
		if err != nil {
			log.Println("[ZRPC] Response Error ->", err)
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
		log.Println("[ZRPC] Response Error ->", err)
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
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"services": s,
		})
		if err != nil {
			log.Println("[ZRPC] Response Error ->", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "GET" {
		if !proxy.ui {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.HasPrefix(r.URL.EscapedPath(), "/ui") {
			proxy.WebUI(w, r)
			return
		}

		http.Redirect(w, r, "/ui", http.StatusTemporaryRedirect)
		return
	}

	if r.URL.EscapedPath() != proxy.PrefixPath {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var data Input
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		err = json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    "500",
				Message: err.Error(),
				Data:    err,
			},
			ID: data.ID,
		})
		if err != nil {
			log.Println("[ZRPC] Response Error ->", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// 先接收輸入的adress
	var address = data.Address

	// 檢查服務是否存在
	service, ok := proxy.Services[data.Service]
	if !ok {
		err = json.NewEncoder(w).Encode(Output{
			Result: nil,
			Error: ErrorDetail{
				Code:    "500",
				Message: "Service Not Found",
				Data:    "Service: " + data.Service,
			},
			ID: data.ID,
		})
		if err != nil {
			log.Println("[ZRPC] Response Error ->", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// 如果沒有輸入address，取註冊服務的address
	if address == "" {
		address = service.RPCAddress
	}

	if proxy.debug {
		log.Printf("[ZRPC] Server (%s), Redirect to %s", service.Name, address)
	}

	res, err := transferJSONRPCClient(address, data.Method, data.Params)
	if err != nil {
		output := Output{
			Result: nil,
			Error:  nil,
			ID:     data.ID,
		}

		jsonrpcErr, yes := IsZrpcError(err)
		if yes {
			output.Error = jsonrpcErr
		} else {
			log.Println("[ZRPC] JSON DECODE Error ->", err)
			output.Error = ErrorDetail{
				Code:    "500",
				Message: err.Error(),
				Data:    err,
			}
		}

		err = json.NewEncoder(w).Encode(output)
		if err != nil {
			log.Println("[ZRPC] Response Error ->", err)
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
		log.Println("[ZRPC] Response Error ->", err)
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
