package zrpc

import (
	"fmt"
	"net/http"
	"strings"
)

// WebUI 顯示介面
func (proxy *Proxy) WebUI(w http.ResponseWriter, r *http.Request) {
	var (
		body string
	)
	// 設定Header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// 查看有沒有服務參數
	name := r.URL.Query().Get("service")

	// 寫網頁
	if s, ok := proxy.Services[name]; ok {
		body = head() + service(s) + footer()
		// body = head() + navbar(r.URL.EscapedPath()) + service(s) + footer()
	} else {
		body = head() + services(proxy.Services) + footer()
		// body = head() + navbar(r.URL.EscapedPath()) + services(proxy.Services) + footer()
	}

	// 輸出網頁
	w.Write([]byte(body))
}

func head() string {
	return `
	<!DOCTYPE html>
	<html>
	<head>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>ZRPC Web</title>
	<style>
		table {
			font-family: arial, sans-serif;
			border-collapse: collapse;
			width: 100%;
		}

		td, th {
			border: 1px solid #dddddd;
			text-align: left;
			padding: 8px;
		}

		tr:nth-child(even) {
			background-color: #dddddd;
		}
		/* Navbar container */
		.navbar {
			overflow: hidden;
			background-color: #333;
			font-family: Arial;
		}

		/* Links inside the navbar */
		.navbar a {
			float: left;
			font-size: 16px;
			color: white;
			text-align: center;
			padding: 14px 16px;
			text-decoration: none;
		}

		/* Links inside the navbar */
		.navbar a.active {
			background-color: red;
		}

		/* The dropdown container */
		.dropdown {
			float: left;
			overflow: hidden;
		}

		/* Dropdown button */
		.dropdown .dropbtn {
			font-size: 16px; 
			border: none;
			outline: none;
			color: white;
			padding: 14px 16px;
			background-color: inherit;
			font-family: inherit; /* Important for vertical align on mobile phones */
			margin: 0; /* Important for vertical align on mobile phones */
		}

		/* Add a red background color to navbar links on hover */
		.navbar a:hover, .dropdown:hover .dropbtn {
			background-color: powderblue;
			color: black;
		}

		/* Dropdown content (hidden by default) */
		.dropdown-content {
			display: none;
			position: absolute;
			background-color: #f9f9f9;
			min-width: 160px;
			box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
			z-index: 1;
		}

		/* Links inside the dropdown */
		.dropdown-content a {
			float: none;
			color: black;
			padding: 12px 16px;
			text-decoration: none;
			display: block;
			text-align: left;
		}

		/* Add a grey background color to dropdown links on hover */
		.dropdown-content a:hover {
			background-color: #ddd;
		}

		/* Show the dropdown menu on hover */
		.dropdown:hover .dropdown-content {
			display: block;
		}
	</style>
	</head>
	<body>
	`
}

func navbar(path string) (html string) {
	html = `<div class="navbar">`
	type link struct {
		name string
		url  string
	}
	urls := []link{
		link{
			name: "Registry",
			url:  "/ui",
		},
		// link{
		// 	name: "Call",
		// 	url:  "/ui/call",
		// },
	}
	for _, url := range urls {
		html += fmt.Sprintf(`<a href="%s" `, url.url)
		if url.url == path {
			html += `class="active"`
		}
		html += fmt.Sprintf(`>%s</a>`, url.name)
	}
	html += "</div>"
	return
}

func footer() string {
	return `
	</body>
	</html>
	`
}

func service(service Service) (html string) {
	html = fmt.Sprintf(`
	<h2>%s - 方法清單 <a href="/ui" title="服務清單">[back]</a></h2>
	`, service.Name)

	srvs, err := getService(service.HTTPAddress)
	if err != nil {
		html += "<h3>Internal Error</h3>"
		html += "<h4 style='color:red;'>" + err.Error() + "</h4>"
		return
	}

	html += ""
	for _, s := range srvs {
		html += fmt.Sprintf(`
		<table>
		<tr>
			<td>服務名稱</td>
			<td>%s</td>
		</tr>
		<tr>
			<th>方法</th>
			<th>參數</th>
		</tr>
		`, s.Name)
		for method, params := range s.Methods {
			html += fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%s</td>
			</tr>
			`, method, params)
		}
		html += "</table>"
	}
	return
}

func services(services map[string]Service) (html string) {
	var (
		i        = 1
		rpcAddr  string
		httpAddr string
	)
	html = `
	<h2>服務位址清單</h2>
	<table>
		<tr>
			<th>#</th>
			<th>服務</th>
			<th>TCP 位址</th>
			<th>HTTP 位址</th>
		</tr>
	`

	for name, service := range services {
		if strings.HasPrefix(service.RPCAddress, ":") {
			rpcAddr = "0.0.0.0" + service.RPCAddress
		} else {
			rpcAddr = service.RPCAddress
		}
		if strings.HasPrefix(service.HTTPAddress, ":") {
			httpAddr = "0.0.0.0" + service.HTTPAddress
		} else {
			httpAddr = service.HTTPAddress
		}

		html += fmt.Sprintf(`
		<tr>
			<td>%d</td>
			<td><a href="/ui?service=%s">%s</a></td>
			<td>%s</td>
			<td>%s</td>
		</tr>
		`, i, name, service.Name, rpcAddr, httpAddr)
		i++
	}
	html += "</table>"
	return
}
