package zrpc

import (
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
