package types

// PaginationRequest 定义通用分页请求参数。
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// TraceRequest 定义可选的 TraceID 传递结构。
type TraceRequest struct {
	TraceID string `json:"traceId,omitempty"`
}
