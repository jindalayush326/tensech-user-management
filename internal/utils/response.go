package utils

import "github.com/gin-gonic/gin"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func Success(c *gin.Context, status int, data interface{}) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	c.JSON(status, response)
}

func Error(c *gin.Context, status int, message string) {
	response := ErrorResponse{
		Success: false,
		Error:   message,
	}

	c.JSON(status, response)
}