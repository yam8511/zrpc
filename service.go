package zrpc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

// Service 服務
type Service struct {
	Name        string            `json:"name,omitempty"`
	Methods     map[string]string `json:"methods,omitempty"`
	RPCAddress  string            `json:"rpc_address,omitempty"`
	HTTPAddress string            `json:"http_address,omitempty"`
}

// ReflectMethod 反映服務可用方法
func ReflectMethod(service interface{}) (name string, methods map[string]string) {
	name = reflect.TypeOf(service).String()
	methods = map[string]string{}
	// 遍歷物件中的方法
	for m := 0; m < reflect.TypeOf(service).NumMethod(); m++ {
		method := reflect.TypeOf(service).Method(m)
		methodType := strings.Replace(method.Type.String(), name+", ", "", 1)
		methods[method.Name] = methodType
	}
	name = strings.Replace(name, "*main.", "", 1)
	return
}

func getService(addr string) ([]Service, error) {

	url := "http://" + addr + "/services"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []Service{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Service{}, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []Service{}, err
	}
	var srv []Service
	err = json.Unmarshal(body, &srv)
	if err != nil {
		return []Service{}, err
	}

	return srv, nil
}
