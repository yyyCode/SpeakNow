package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Data: data})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 400, message)
}

func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, 429, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, 500, message)
}

func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, 503, message)
}
