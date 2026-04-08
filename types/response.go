package types

// HealthResponse 定义后续健康检查接口的最小响应结构。
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Stage   string `json:"stage"`
}

// CommandResponse 定义 CLI 或控制面反馈的最小响应结构。
type CommandResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Stage   string `json:"stage"`
}
