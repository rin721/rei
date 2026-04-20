package types

// SampleItemResponse 定义示例业务模块响应。
type SampleItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SampleToolkitDemoResponse describes a business-level toolkit demo.
type SampleToolkitDemoResponse struct {
	Module   string `json:"module"`
	Scenario string `json:"scenario"`
	Guidance string `json:"guidance"`
	Preview  string `json:"preview"`
}
