package zrpc

// Service 服務
type Service struct {
	Name    string            `json:"name,omitempty"`
	Methods map[string]string `json:"methods,omitempty"`
	Address string            `json:"address,omitempty"`
}
