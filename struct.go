package zrpc

import "encoding/json"

// Input 輸出參數
type Input struct {
	Service string      `json:"service"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
	Address string      `json:"address"`
}

// Output 輸出參數
type Output struct {
	Result interface{} `json:"result"`
	Error  error       `json:"error"`
	ID     int         `json:"id"`
}

// ErrorDetail 錯誤細節
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Error 顯示ErrorDetail的訊息
func (e ErrorDetail) Error() string {
	errMsg, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}
	return string(errMsg)
}
