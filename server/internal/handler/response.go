package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const invalidParamsMessage = "鏃犳晥鐨勫弬鏁?"

func respondSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
}

func respondSuccessWithExtras(c *gin.Context, data any, extras gin.H) {
	payload := gin.H{
		"success": true,
		"message": "",
		"data":    data,
	}
	for key, value := range extras {
		payload[key] = value
	}
	c.JSON(http.StatusOK, payload)
}

func respondSuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

func respondFailure(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": message,
	})
}

func respondBadRequest(c *gin.Context, message string) {
	if message == "" {
		message = invalidParamsMessage
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": message,
	})
}

func respondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"message": message,
	})
}

func decodeJSONBody(body io.Reader, target any) error {
	return json.NewDecoder(body).Decode(target)
}

func decodeOptionalJSONBody(body io.Reader, target any) error {
	if err := json.NewDecoder(body).Decode(target); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
