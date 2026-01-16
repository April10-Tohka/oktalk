package response

import (
	"net/http"
	"oktalk/internal/pkg/constants"

	"github.com/gin-gonic/gin"
)

// Response 统一 JSON 结构
type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id"`
}

func SendJSON(c *gin.Context, businessCode int, data interface{}, msg string) {
	traceID := c.GetString(constants.TraceIDKey)

	c.JSON(http.StatusOK, Response{
		TraceID: traceID,
		Code:    businessCode,
		Msg:     msg,
		Data:    data,
	})
}
