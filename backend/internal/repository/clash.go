package repository

import (
	"context"

	"github.com/Rain-kl/Foam/backend/internal/domain/clash"
)

type SourceConfigRepository interface {
	Create(ctx context.Context, item *clash.SourceConfig) error
	GetByID(ctx context.Context, id int) (*clash.SourceConfig, error)
	GetByHash(ctx context.Context, hash string) (*clash.SourceConfig, error)
	List(ctx context.Context, page, pageSize int) ([]*clash.SourceConfig, int64, error)
	Update(ctx context.Context, item *clash.SourceConfig) error
	Delete(ctx context.Context, id int) error
}

type ProxyNodeRepository interface {
	Create(ctx context.Context, item *clash.ProxyNode) error
	CreateBatch(ctx context.Context, items []*clash.ProxyNode) error
	GetByID(ctx context.Context, id int) (*clash.ProxyNode, error)
	GetByIDs(ctx context.Context, ids []int) ([]*clash.ProxyNode, error)
	FindExistingFingerprints(ctx context.Context, fingerprints []string) (map[string]struct{}, error)
	FindExistingFingerprintsExcludingSource(ctx context.Context, fingerprints []string, excludeSourceID int) (map[string]struct{}, error)
	List(ctx context.Context, page, pageSize int, filter clash.ProxyNodeFilter) ([]*clash.ProxyNode, int64, error)
	ListIDsBySourceConfigID(ctx context.Context, sourceConfigID int) ([]int, error)
	ListBySourceConfigID(ctx context.Context, sourceConfigID int) ([]*clash.ProxyNode, error)
	Update(ctx context.Context, item *clash.ProxyNode) error
	UpdateBatch(ctx context.Context, items []*clash.ProxyNode) error
	UpdateTestStatus(ctx context.Context, id int, status string, latencyMS *int, testErr string) error
	Delete(ctx context.Context, id int) error
	DeleteBatch(ctx context.Context, ids []int) error
	// DeleteBySourceConfigID removes all nodes for a source and their
	// port_profile_nodes / node_test_results bindings in one transaction.
	DeleteBySourceConfigID(ctx context.Context, sourceConfigID int) error
	ToggleBatch(ctx context.Context, ids []int, enabled bool) error
}

type NodeTestResultRepository interface {
	Create(ctx context.Context, item *clash.NodeTestResult) error
	ListByNodeID(ctx context.Context, nodeID int, limit int) ([]*clash.NodeTestResult, error)
}

type PortProfileRepository interface {
	Create(ctx context.Context, item *clash.PortProfile) error
	GetByID(ctx context.Context, id int) (*clash.PortProfile, error)
	List(ctx context.Context) ([]*clash.PortProfile, error)
	Update(ctx context.Context, item *clash.PortProfile) error
	Delete(ctx context.Context, id int) error

	// Node bindings
	SetProfileNodes(ctx context.Context, profileID int, nodeIDs []int) error
	GetProfileNodeIDs(ctx context.Context, profileID int) ([]int, error)
	GetProfileNodes(ctx context.Context, profileID int) ([]*clash.ProxyNode, error)
}

type PortProfileTemplateRepository interface {
	Create(ctx context.Context, item *clash.PortProfileTemplate) error
	GetByID(ctx context.Context, id int) (*clash.PortProfileTemplate, error)
	List(ctx context.Context) ([]*clash.PortProfileTemplate, error)
	Delete(ctx context.Context, id int) error
}

type RuntimeConfigRepository interface {
	Upsert(ctx context.Context, item *clash.RuntimeConfig) error
	GetByPortProfileID(ctx context.Context, profileID int) (*clash.RuntimeConfig, error)
	ListAll(ctx context.Context) ([]*clash.RuntimeConfig, error)
	DeleteByPortProfileID(ctx context.Context, profileID int) error
}

type KernelInstanceRepository interface {
	Upsert(ctx context.Context, item *clash.KernelInstance) error
	GetByType(ctx context.Context, kernelType string) (*clash.KernelInstance, error)
}
