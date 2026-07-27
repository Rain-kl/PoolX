package example

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	exampleapp "github.com/Rain-kl/Foam/backend/internal/application/example"
	"github.com/Rain-kl/Foam/backend/internal/infra/persistence/relational"
	"github.com/gin-gonic/gin"
)

func TestExampleHTTPCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)

	database, err := relational.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "example-http.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.InitializeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}

	service := exampleapp.NewService(relational.NewExampleRepository(database))
	handler := NewHandler(service)
	router := gin.New()
	api := router.Group("/api/v1/admin")
	handler.Register(api)

	createBody := []byte(`{"name":"demo","description":"scaffold"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/examples", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", createRec.Code, createRec.Body.String())
	}

	var createResp struct {
		ErrorMsg string `json:"error_msg"`
		Data     struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatal(err)
	}
	if createResp.ErrorMsg != "" || createResp.Data.ID == "" || createResp.Data.Name != "demo" {
		t.Fatalf("create response = %#v", createResp)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/examples?page=1&page_size=10", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d body = %s", listRec.Code, listRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/examples/"+createResp.Data.ID, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d body = %s", getRec.Code, getRec.Body.String())
	}

	updateBody := []byte(`{"name":"demo-2","description":"updated"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/admin/examples/"+createResp.Data.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d body = %s", updateRec.Code, updateRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/examples/"+createResp.Data.ID, nil)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d body = %s", deleteRec.Code, deleteRec.Body.String())
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/examples/"+createResp.Data.ID, nil)
	missingRec := httptest.NewRecorder()
	router.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d body = %s", missingRec.Code, missingRec.Body.String())
	}
	if !strings.Contains(missingRec.Body.String(), `"error_msg"`) || !strings.Contains(missingRec.Body.String(), "不存在") {
		t.Fatalf("missing body = %s", missingRec.Body.String())
	}
}
