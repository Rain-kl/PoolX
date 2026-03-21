package router_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"poolx/internal/model"
	"poolx/internal/pkg/common"
)

type apiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func setupTestDB(t *testing.T) {
	t.Helper()

	originalSQLitePath := common.SQLitePath

	common.SQLitePath = filepath.Join(t.TempDir(), "router-test.db")

	if err := model.InitDB(); err != nil {
		t.Fatalf("init test db: %v", err)
	}

	t.Cleanup(func() {
		if err := model.CloseDB(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
		model.DB = nil
		common.SQLitePath = originalSQLitePath
	})
}
