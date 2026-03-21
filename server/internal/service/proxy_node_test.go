package service

import (
	"context"
	"testing"

	"poolx/internal/model"
	"poolx/internal/pkg/common"
	kernelpkg "poolx/internal/pkg/kernel"
)

func TestExecuteNodeTestsPersistsResult(t *testing.T) {
	setupServiceTestDB(t)

	originalRunner := runNodeKernelTest
	runNodeKernelTest = func(ctx context.Context, input kernelpkg.MihomoNodeTestInput) (*kernelpkg.MihomoNodeTestResult, error) {
		return &kernelpkg.MihomoNodeTestResult{LatencyMS: 321}, nil
	}
	t.Cleanup(func() {
		runNodeKernelTest = originalRunner
	})

	originalBinaryPath := common.MihomoBinaryPath
	common.MihomoBinaryPath = "/tmp/fake-mihomo"
	t.Cleanup(func() {
		common.MihomoBinaryPath = originalBinaryPath
	})

	node := &model.ProxyNode{
		SourceConfigID:   1,
		SourceConfigName: "seed.yaml",
		Name:             "local-node",
		Type:             "ss",
		Server:           "127.0.0.1",
		Port:             1,
		Fingerprint:      "fingerprint-local-node",
		MetadataJSON:     `{"name":"local-node"}`,
		Enabled:          true,
		LastTestStatus:   model.NodeTestStatusUnknown,
	}
	if err := model.DB.Create(node).Error; err != nil {
		t.Fatalf("seed proxy node: %v", err)
	}

	results, err := ExecuteNodeTests(context.Background(), NodeTestInput{
		NodeIDs:   []int{node.ID},
		TimeoutMS: 1000,
	})
	if err != nil {
		t.Fatalf("ExecuteNodeTests returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one test result, got %d", len(results))
	}
	if results[0].Status != model.NodeTestStatusSuccess {
		t.Fatalf("unexpected test status: %+v", results[0])
	}

	var refreshed model.ProxyNode
	if err := model.DB.First(&refreshed, "id = ?", node.ID).Error; err != nil {
		t.Fatalf("reload proxy node: %v", err)
	}
	if refreshed.LastTestStatus != model.NodeTestStatusSuccess {
		t.Fatalf("expected node status to be updated, got %s", refreshed.LastTestStatus)
	}

	rows, err := model.ListNodeTestResults(node.ID, 10)
	if err != nil {
		t.Fatalf("ListNodeTestResults returned error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one persisted node test row, got %d", len(rows))
	}
	if rows[0].LatencyMS == nil || *rows[0].LatencyMS != 321 {
		t.Fatalf("expected persisted latency to match kernel test result: %+v", rows[0])
	}
}

func TestExecuteNodeTestsPersistsFailureResult(t *testing.T) {
	setupServiceTestDB(t)

	originalRunner := runNodeKernelTest
	runNodeKernelTest = func(ctx context.Context, input kernelpkg.MihomoNodeTestInput) (*kernelpkg.MihomoNodeTestResult, error) {
		return nil, assertiveError("boom")
	}
	t.Cleanup(func() {
		runNodeKernelTest = originalRunner
	})

	originalBinaryPath := common.MihomoBinaryPath
	common.MihomoBinaryPath = "/tmp/fake-mihomo"
	t.Cleanup(func() {
		common.MihomoBinaryPath = originalBinaryPath
	})

	node := &model.ProxyNode{
		SourceConfigID:   1,
		SourceConfigName: "seed.yaml",
		Name:             "local-node",
		Type:             "ss",
		Server:           "127.0.0.1",
		Port:             1,
		Fingerprint:      "fingerprint-local-node-failed",
		MetadataJSON:     `{"name":"local-node"}`,
		Enabled:          true,
		LastTestStatus:   model.NodeTestStatusUnknown,
	}
	if err := model.DB.Create(node).Error; err != nil {
		t.Fatalf("seed proxy node: %v", err)
	}

	results, err := ExecuteNodeTests(context.Background(), NodeTestInput{
		NodeIDs: []int{node.ID},
	})
	if err != nil {
		t.Fatalf("ExecuteNodeTests returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one test result, got %d", len(results))
	}
	if results[0].Status != model.NodeTestStatusFailed {
		t.Fatalf("unexpected test status: %+v", results[0])
	}

	var refreshed model.ProxyNode
	if err := model.DB.First(&refreshed, "id = ?", node.ID).Error; err != nil {
		t.Fatalf("reload proxy node: %v", err)
	}
	if refreshed.LastTestStatus != model.NodeTestStatusFailed {
		t.Fatalf("expected node status to be updated, got %s", refreshed.LastTestStatus)
	}

	rows, err := model.ListNodeTestResults(node.ID, 10)
	if err != nil {
		t.Fatalf("ListNodeTestResults returned error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one persisted node test row, got %d", len(rows))
	}
}

type assertiveError string

func (e assertiveError) Error() string {
	return string(e)
}
