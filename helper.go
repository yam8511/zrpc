package zrpc

import (
	"encoding/json"
	"log"
	"os"
)

// 偵測訊號
func (server *Server) delectSignal(done bool, sig chan os.Signal, prevSig os.Signal) (os.Signal, bool) {
	if done {
		return prevSig, true
	}
	select {
	case s := <-sig:
		return s, true
	default:
		return nil, false
	}
}

// 等待連線結束
func (server *Server) waitConnection(done bool, sig chan os.Signal, prevSig os.Signal, e chan error) (bool, os.Signal, error) {
	if done {
		select {
		case ip := <-server.rpcIn:
			server.online++
			if server.debug {
				log.Println("CONNECT ->", ip)
			}
		case ip := <-server.httpIn:
			server.online++
			if server.debug {
				log.Println("CONNECT ->", ip)
			}
		case ip := <-server.rpcOut:
			server.online--
			if server.debug {
				log.Println("DISCONNECT ->", ip)
			}
		case ip := <-server.httpOut:
			server.online--
			if server.debug {
				log.Println("DISCONNECT ->", ip)
			}
		case err := <-e:
			return false, prevSig, err
		}
	} else {
		select {
		case ip := <-server.rpcIn:
			server.online++
			if server.debug {
				log.Println("CONNECT ->", ip)
			}
		case ip := <-server.httpIn:
			server.online++
			if server.debug {
				log.Println("CONNECT ->", ip)
			}
		case ip := <-server.rpcOut:
			server.online--
			if server.debug {
				log.Println("DISCONNECT ->", ip)
			}
		case ip := <-server.httpOut:
			server.online--
			if server.debug {
				log.Println("DISCONNECT ->", ip)
			}
		case s := <-sig:
			return true, s, nil
		case err := <-e:
			return false, prevSig, err
		}
	}
	return done, prevSig, nil
}

// IsZrpcError 是否為套件的錯誤型態
func IsZrpcError(e error) (detail *ErrorDetail, yes bool) {
	err := json.Unmarshal([]byte(e.Error()), &detail)
	yes = err == nil
	return
}

// NewZrpcError 建立ZRPC的錯誤
func NewZrpcError(code, message string, data interface{}) *ErrorDetail {
	return &ErrorDetail{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
