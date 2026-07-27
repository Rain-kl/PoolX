package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlerReturnsStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(func() string { return "https://api.example.com/" }).Register(router.Group("/api/v1/admin"))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/status", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var payload struct {
		ErrorMsg string         `json:"error_msg"`
		Data     map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ErrorMsg != "" {
		t.Fatalf("error_msg = %q", payload.ErrorMsg)
	}
	if payload.Data["public_api_base_url"] != "https://api.example.com" {
		t.Fatalf("data = %#v", payload.Data)
	}
	if payload.Data["ok"] != true {
		t.Fatalf("data.ok = %#v", payload.Data["ok"])
	}
}

func TestHandlerReturnsVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(nil).Register(router.Group("/api/v1/admin"))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/version", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		ErrorMsg string         `json:"error_msg"`
		Data     map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ErrorMsg != "" || payload.Data["current_version"] == nil {
		t.Fatalf("payload = %#v", payload)
	}
	if _, ok := payload.Data["is_canary"]; !ok {
		t.Fatalf("missing is_canary: %#v", payload.Data)
	}
}
