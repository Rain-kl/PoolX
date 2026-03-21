package router_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"poolx/internal/pkg/common"
	"poolx/internal/router"
	"poolx/internal/service"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLatestReleaseProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	common.RedisEnabled = false
	setupTestDB(t)
	originalRepo := common.ServerUpdateRepo
	common.ServerUpdateRepo = "example/custom-template"
	t.Cleanup(func() {
		common.ServerUpdateRepo = originalRepo
	})

	originalClient := service.UpdateHTTPClientForTest()
	service.SetUpdateHTTPClientForTest(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.github.com/repos/example/custom-template/releases/latest" {
				t.Fatalf("unexpected request url: %s", req.URL.String())
			}
			if req.Header.Get("Accept") != "application/vnd.github+json" {
				t.Fatalf("unexpected accept header: %s", req.Header.Get("Accept"))
			}
			if req.Header.Get("User-Agent") != "PoolX-Server" {
				t.Fatalf("unexpected user-agent header: %s", req.Header.Get("User-Agent"))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body: io.NopCloser(strings.NewReader(`{
					"tag_name":"v1.2.3",
					"body":"release notes",
					"html_url":"https://github.com/Rain-kl/PoolX/releases/tag/v1.2.3",
					"published_at":"2026-03-11T00:00:00Z"
				}`)),
			}, nil
		}),
	})
	t.Cleanup(func() {
		service.SetUpdateHTTPClientForTest(originalClient)
	})

	engine := gin.New()
	engine.Use(sessions.Sessions("session", cookie.NewStore([]byte("test-secret"))))
	router.SetApiRouter(engine)

	loginBody, err := json.Marshal(map[string]string{
		"username": "root",
		"password": "123456",
	})
	if err != nil {
		t.Fatalf("failed to marshal login body: %v", err)
	}
	loginReq := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	engine.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected login status code: %d", loginRecorder.Code)
	}
	loginResult := loginRecorder.Result()
	defer loginResult.Body.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/update/latest-release", nil)
	for _, cookieValue := range loginResult.Cookies() {
		req.AddCookie(cookieValue)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}

	var resp apiResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success response, got message: %s", resp.Message)
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to decode response data: %v", err)
	}
	if data["tag_name"] != "v1.2.3" {
		t.Fatalf("unexpected tag_name: %#v", data["tag_name"])
	}
	if data["current_version"] != common.Version {
		t.Fatalf("unexpected current_version: %#v", data["current_version"])
	}
}

func loginRootAndBuildEngine(t *testing.T) (*gin.Engine, []*http.Cookie) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.RedisEnabled = false
	setupTestDB(t)

	engine := gin.New()
	engine.Use(sessions.Sessions("session", cookie.NewStore([]byte("test-secret"))))
	router.SetApiRouter(engine)

	loginBody, err := json.Marshal(map[string]string{
		"username": "root",
		"password": "123456",
	})
	if err != nil {
		t.Fatalf("failed to marshal login body: %v", err)
	}
	loginReq := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	engine.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected login status code: %d", loginRecorder.Code)
	}
	loginResult := loginRecorder.Result()
	defer loginResult.Body.Close()

	return engine, loginResult.Cookies()
}

func fakeManualServerBinary(version string) (string, []byte) {
	if runtime.GOOS == "windows" {
		return "poolx-server-test.cmd", []byte("@echo off\r\necho " + version + "\r\n")
	}
	return "poolx-server-test.sh", []byte("#!/bin/sh\necho " + version + "\n")
}

func TestManualUploadRoute(t *testing.T) {
	originalVersion := common.Version
	common.Version = "v0.4.0"
	t.Cleanup(func() {
		common.Version = originalVersion
		service.SetServerBinaryUpgradeExecutorForTest(nil)
		service.SetServerUpgradeDispatchDelayForTest(500 * time.Millisecond)
	})

	engine, cookies := loginRootAndBuildEngine(t)
	fileName, content := fakeManualServerBinary("v0.5.0")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("binary", fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatalf("failed to write upload content: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/update/manual-upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookieValue := range cookies {
		req.AddCookie(cookieValue)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}

	var resp apiResponse
	if err = json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success response, got message: %s", resp.Message)
	}

	var data map[string]any
	if err = json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to decode response data: %v", err)
	}
	if data["detected_version"] != "v0.5.0" {
		t.Fatalf("unexpected detected_version: %#v", data["detected_version"])
	}
	if data["ready_to_upgrade"] != true {
		t.Fatalf("expected ready_to_upgrade to be true: %#v", data["ready_to_upgrade"])
	}
	if data["upload_token"] == "" {
		t.Fatal("expected upload_token to be returned")
	}
}

func TestManualUpgradeConfirmRoute(t *testing.T) {
	originalVersion := common.Version
	originalExecutor := service.ServerBinaryUpgradeExecutorForTest()
	originalDelay := service.ServerUpgradeDispatchDelayForTest()
	common.Version = "v0.4.0"
	called := make(chan string, 1)
	service.SetServerBinaryUpgradeExecutorForTest(func(execPath string, tempPath string) error {
		called <- tempPath
		return nil
	})
	service.SetServerUpgradeDispatchDelayForTest(0)
	t.Cleanup(func() {
		common.Version = originalVersion
		service.SetServerBinaryUpgradeExecutorForTest(originalExecutor)
		service.SetServerUpgradeDispatchDelayForTest(originalDelay)
	})

	engine, cookies := loginRootAndBuildEngine(t)
	fileName, content := fakeManualServerBinary("v0.5.0")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("binary", fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatalf("failed to write upload content: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	uploadReq := httptest.NewRequest(http.MethodPost, "/api/update/manual-upload", body)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookieValue := range cookies {
		uploadReq.AddCookie(cookieValue)
	}

	uploadRecorder := httptest.NewRecorder()
	engine.ServeHTTP(uploadRecorder, uploadReq)
	if uploadRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected upload status code: %d", uploadRecorder.Code)
	}

	var uploadResp apiResponse
	if err = json.Unmarshal(uploadRecorder.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("failed to decode upload response: %v", err)
	}
	if !uploadResp.Success {
		t.Fatalf("expected upload success, got message: %s", uploadResp.Message)
	}

	var uploadData map[string]any
	if err = json.Unmarshal(uploadResp.Data, &uploadData); err != nil {
		t.Fatalf("failed to decode upload response data: %v", err)
	}
	uploadToken, _ := uploadData["upload_token"].(string)
	if uploadToken == "" {
		t.Fatal("expected upload token in upload response")
	}

	confirmBody, err := json.Marshal(map[string]string{"upload_token": uploadToken})
	if err != nil {
		t.Fatalf("failed to marshal confirm body: %v", err)
	}
	confirmReq := httptest.NewRequest(http.MethodPost, "/api/update/manual-upgrade", bytes.NewReader(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	for _, cookieValue := range cookies {
		confirmReq.AddCookie(cookieValue)
	}

	confirmRecorder := httptest.NewRecorder()
	engine.ServeHTTP(confirmRecorder, confirmReq)
	if confirmRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected confirm status code: %d", confirmRecorder.Code)
	}

	var confirmResp apiResponse
	if err = json.Unmarshal(confirmRecorder.Body.Bytes(), &confirmResp); err != nil {
		t.Fatalf("failed to decode confirm response: %v", err)
	}
	if !confirmResp.Success {
		t.Fatalf("expected confirm success, got message: %s", confirmResp.Message)
	}

	select {
	case tempPath := <-called:
		if tempPath == "" {
			t.Fatal("expected manual upgrade executor to receive temp path")
		}
	case <-time.After(time.Second):
		t.Fatal("expected manual upgrade executor to be called")
	}
}
