package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rin721/rei/internal/service"
)

// SampleHandler 负责示例业务模块接口绑定与响应。
type SampleHandler struct {
	service service.SampleService
}

// NewSampleHandler 创建示例处理器。
func NewSampleHandler(svc service.SampleService) *SampleHandler {
	return &SampleHandler{service: svc}
}

// List 返回示例数据列表。
func (h *SampleHandler) List(c *gin.Context) {
	response, err := h.service.List(c.Request.Context())
	if err != nil {
		writeFailure(c, statusFromError(err), err)
		return
	}

	writeSuccess(c, http.StatusOK, response)
}
