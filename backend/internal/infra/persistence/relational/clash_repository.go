package relational

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/domain/clash"
	"github.com/Rain-kl/Foam/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SourceConfigRepositoryImpl
type SourceConfigRepositoryImpl struct {
	db *gorm.DB
}

func NewSourceConfigRepository(db *Database) repository.SourceConfigRepository {
	return &SourceConfigRepositoryImpl{db: db.db}
}

func (r *SourceConfigRepositoryImpl) Create(ctx context.Context, item *clash.SourceConfig) error {
	m := toSourceConfigModel(item)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	item.ID = m.ID
	item.CreatedAt = m.CreatedAt
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *SourceConfigRepositoryImpl) GetByID(ctx context.Context, id int) (*clash.SourceConfig, error) {
	var m sourceConfigModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return toSourceConfigDomain(&m), nil
}

func (r *SourceConfigRepositoryImpl) GetByHash(ctx context.Context, hash string) (*clash.SourceConfig, error) {
	var m sourceConfigModel
	if err := r.db.WithContext(ctx).First(&m, "content_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return toSourceConfigDomain(&m), nil
}

func (r *SourceConfigRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*clash.SourceConfig, int64, error) {
	page, pageSize = repository.NormalizePage(page, pageSize, repository.DefaultPageSize)
	offset := (page - 1) * pageSize

	var total int64
	if err := r.db.WithContext(ctx).Model(&sourceConfigModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []sourceConfigModel
	if err := r.db.WithContext(ctx).Order("id desc").Limit(pageSize).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*clash.SourceConfig, len(models))
	for i, m := range models {
		items[i] = toSourceConfigDomain(&m)
	}
	return items, total, nil
}

func (r *SourceConfigRepositoryImpl) Update(ctx context.Context, item *clash.SourceConfig) error {
	m := toSourceConfigModel(item)
	m.UpdatedAt = time.Now()
	// Select all mutable columns so zero values (e.g. imported_nodes=0) are persisted.
	if err := r.db.WithContext(ctx).Model(&sourceConfigModel{}).Where("id = ?", item.ID).
		Select(
			"source_type", "source_url", "content_type", "fetched_at", "filename",
			"content_hash", "raw_content", "status",
			"total_nodes", "valid_nodes", "invalid_nodes", "duplicate_nodes", "imported_nodes",
			"uploaded_by", "uploaded_by_id", "updated_at",
		).
		Updates(m).Error; err != nil {
		return err
	}
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *SourceConfigRepositoryImpl) Delete(ctx context.Context, id int) error {
	res := r.db.WithContext(ctx).Delete(&sourceConfigModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// ProxyNodeRepositoryImpl
type ProxyNodeRepositoryImpl struct {
	db *gorm.DB
}

func NewProxyNodeRepository(db *Database) repository.ProxyNodeRepository {
	return &ProxyNodeRepositoryImpl{db: db.db}
}

func (r *ProxyNodeRepositoryImpl) Create(ctx context.Context, item *clash.ProxyNode) error {
	m := toProxyNodeModel(item)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	item.ID = m.ID
	item.CreatedAt = m.CreatedAt
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *ProxyNodeRepositoryImpl) CreateBatch(ctx context.Context, items []*clash.ProxyNode) error {
	if len(items) == 0 {
		return nil
	}
	models := make([]*proxyNodeModel, len(items))
	now := time.Now()
	for i, item := range items {
		m := toProxyNodeModel(item)
		if m.CreatedAt.IsZero() {
			m.CreatedAt = now
		}
		m.UpdatedAt = now
		models[i] = m
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(models, 100).Error; err != nil {
		return err
	}
	for i, m := range models {
		items[i].ID = m.ID
		items[i].CreatedAt = m.CreatedAt
		items[i].UpdatedAt = m.UpdatedAt
	}
	return nil
}

func (r *ProxyNodeRepositoryImpl) GetByID(ctx context.Context, id int) (*clash.ProxyNode, error) {
	var m proxyNodeModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return toProxyNodeDomain(&m), nil
}

func (r *ProxyNodeRepositoryImpl) GetByIDs(ctx context.Context, ids []int) ([]*clash.ProxyNode, error) {
	if len(ids) == 0 {
		return []*clash.ProxyNode{}, nil
	}
	var models []proxyNodeModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Order("id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*clash.ProxyNode, len(models))
	for i, m := range models {
		items[i] = toProxyNodeDomain(&m)
	}
	return items, nil
}

func (r *ProxyNodeRepositoryImpl) FindExistingFingerprints(ctx context.Context, fingerprints []string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(fingerprints) == 0 {
		return result, nil
	}
	var rows []string
	if err := r.db.WithContext(ctx).Model(&proxyNodeModel{}).Where("fingerprint IN ?", fingerprints).Pluck("fingerprint", &rows).Error; err != nil {
		return nil, err
	}
	for _, f := range rows {
		result[f] = struct{}{}
	}
	return result, nil
}

func (r *ProxyNodeRepositoryImpl) FindExistingFingerprintsExcludingSource(ctx context.Context, fingerprints []string, excludeSourceID int) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(fingerprints) == 0 {
		return result, nil
	}
	var rows []string
	if err := r.db.WithContext(ctx).Model(&proxyNodeModel{}).
		Where("fingerprint IN ? AND source_config_id != ?", fingerprints, excludeSourceID).
		Pluck("fingerprint", &rows).Error; err != nil {
		return nil, err
	}
	for _, f := range rows {
		result[f] = struct{}{}
	}
	return result, nil
}

func (r *ProxyNodeRepositoryImpl) List(ctx context.Context, page, pageSize int, filter clash.ProxyNodeFilter) ([]*clash.ProxyNode, int64, error) {
	page, pageSize = repository.NormalizePage(page, pageSize, repository.DefaultPageSize)
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&proxyNodeModel{})
	if filter.Keyword != "" {
		kw := filter.Keyword + "%"
		query = query.Where("name LIKE ? OR server LIKE ? OR type LIKE ?", kw, kw, kw)
	}
	if filter.SourceConfigID > 0 {
		query = query.Where("source_config_id = ?", filter.SourceConfigID)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []proxyNodeModel
	if err := query.Order("id desc").Limit(pageSize).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*clash.ProxyNode, len(models))
	for i, m := range models {
		items[i] = toProxyNodeDomain(&m)
	}
	return items, total, nil
}

func (r *ProxyNodeRepositoryImpl) Update(ctx context.Context, item *clash.ProxyNode) error {
	m := toProxyNodeModel(item)
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Model(&proxyNodeModel{}).Where("id = ?", item.ID).Updates(m).Error; err != nil {
		return err
	}
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *ProxyNodeRepositoryImpl) UpdateBatch(ctx context.Context, items []*clash.ProxyNode) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, item := range items {
			m := toProxyNodeModel(item)
			m.UpdatedAt = now
			if err := tx.Model(&proxyNodeModel{}).Where("id = ?", item.ID).
				Select("source_config_name", "name", "type", "server", "port", "metadata_json", "updated_at").
				Updates(m).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProxyNodeRepositoryImpl) UpdateTestStatus(ctx context.Context, id int, status string, latencyMS *int, testErr string) error {
	now := time.Now()
	updates := map[string]any{
		"last_test_status": status,
		"last_latency_ms":  latencyMS,
		"last_test_error":  testErr,
		"last_tested_at":   now,
		"updated_at":       now,
	}
	return r.db.WithContext(ctx).Model(&proxyNodeModel{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ProxyNodeRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&proxyNodeModel{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return repository.ErrNotFound
		}
		return cascadeDeleteProxyNodes(tx, []int{id})
	})
}

func (r *ProxyNodeRepositoryImpl) DeleteBatch(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return cascadeDeleteProxyNodes(tx, ids)
	})
}

func (r *ProxyNodeRepositoryImpl) ListIDsBySourceConfigID(ctx context.Context, sourceConfigID int) ([]int, error) {
	var ids []int
	if err := r.db.WithContext(ctx).Model(&proxyNodeModel{}).
		Where("source_config_id = ?", sourceConfigID).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *ProxyNodeRepositoryImpl) ListBySourceConfigID(ctx context.Context, sourceConfigID int) ([]*clash.ProxyNode, error) {
	var models []proxyNodeModel
	if err := r.db.WithContext(ctx).Where("source_config_id = ?", sourceConfigID).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*clash.ProxyNode, len(models))
	for i, m := range models {
		items[i] = toProxyNodeDomain(&m)
	}
	return items, nil
}

func (r *ProxyNodeRepositoryImpl) DeleteBySourceConfigID(ctx context.Context, sourceConfigID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return cascadeDeleteProxyNodesBySource(tx, sourceConfigID)
	})
}

func cascadeDeleteProxyNodesBySource(tx *gorm.DB, sourceConfigID int) error {
	var ids []int
	if err := tx.Model(&proxyNodeModel{}).
		Where("source_config_id = ?", sourceConfigID).
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	return cascadeDeleteProxyNodes(tx, ids)
}

func cascadeDeleteProxyNodes(tx *gorm.DB, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	if err := tx.Delete(&nodeTestResultModel{}, "node_id IN ?", ids).Error; err != nil {
		return err
	}
	return tx.Delete(&proxyNodeModel{}, "id IN ?", ids).Error
}

func (r *ProxyNodeRepositoryImpl) ToggleBatch(ctx context.Context, ids []int, enabled bool) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&proxyNodeModel{}).Where("id IN ?", ids).Updates(map[string]any{
		"enabled":    enabled,
		"updated_at": time.Now(),
	}).Error
}

// NodeTestResultRepositoryImpl
type NodeTestResultRepositoryImpl struct {
	db *gorm.DB
}

func NewNodeTestResultRepository(db *Database) repository.NodeTestResultRepository {
	return &NodeTestResultRepositoryImpl{db: db.db}
}

func (r *NodeTestResultRepositoryImpl) Create(ctx context.Context, item *clash.NodeTestResult) error {
	m := &nodeTestResultModel{
		NodeID:       item.NodeID,
		TestType:     item.TestType,
		Success:      item.Success,
		LatencyMS:    item.LatencyMS,
		ErrorMessage: item.ErrorMessage,
		TestedAt:     item.TestedAt,
	}
	if m.TestedAt.IsZero() {
		m.TestedAt = time.Now()
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	item.ID = m.ID
	item.TestedAt = m.TestedAt
	return nil
}

func (r *NodeTestResultRepositoryImpl) ListByNodeID(ctx context.Context, nodeID int, limit int) ([]*clash.NodeTestResult, error) {
	if limit <= 0 {
		limit = 10
	}
	var models []nodeTestResultModel
	if err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).Order("id desc").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*clash.NodeTestResult, len(models))
	for i, m := range models {
		items[i] = toNodeTestResultDomain(&m)
	}
	return items, nil
}

// PortProfileRepositoryImpl
type PortProfileRepositoryImpl struct {
	db *gorm.DB
}

func NewPortProfileRepository(db *Database) repository.PortProfileRepository {
	return &PortProfileRepositoryImpl{db: db.db}
}

func (r *PortProfileRepositoryImpl) Create(ctx context.Context, item *clash.PortProfile) error {
	m := toPortProfileModel(item)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	item.ID = m.ID
	item.CreatedAt = m.CreatedAt
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *PortProfileRepositoryImpl) GetByID(ctx context.Context, id int) (*clash.PortProfile, error) {
	var m portProfileModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return toPortProfileDomain(&m), nil
}

func (r *PortProfileRepositoryImpl) List(ctx context.Context) ([]*clash.PortProfile, error) {
	var models []portProfileModel
	if err := r.db.WithContext(ctx).Order("id desc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*clash.PortProfile, len(models))
	for i, m := range models {
		items[i] = toPortProfileDomain(&m)
	}
	return items, nil
}

func (r *PortProfileRepositoryImpl) Update(ctx context.Context, item *clash.PortProfile) error {
	m := toPortProfileModel(item)
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Model(&portProfileModel{}).Where("id = ?", item.ID).Select("*").Updates(m).Error; err != nil {
		return err
	}
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *PortProfileRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_ = tx.Delete(&portProfileNodeModel{}, "port_profile_id = ?", id).Error
		_ = tx.Delete(&runtimeConfigModel{}, "port_profile_id = ?", id).Error
		res := tx.Delete(&portProfileModel{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return repository.ErrNotFound
		}
		return nil
	})
}

func (r *PortProfileRepositoryImpl) SetProfileNodes(ctx context.Context, profileID int, nodeIDs []int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&portProfileNodeModel{}, "port_profile_id = ?", profileID).Error; err != nil {
			return err
		}
		if len(nodeIDs) == 0 {
			return nil
		}

		var nodes []proxyNodeModel
		if err := tx.Where("id IN ?", nodeIDs).Find(&nodes).Error; err != nil {
			return err
		}
		fpMap := make(map[int]string, len(nodes))
		for _, n := range nodes {
			fpMap[n.ID] = n.Fingerprint
		}

		now := time.Now()
		bindings := make([]*portProfileNodeModel, len(nodeIDs))
		for i, nodeID := range nodeIDs {
			bindings[i] = &portProfileNodeModel{
				PortProfileID:   profileID,
				ProxyNodeID:     nodeID,
				NodeFingerprint: fpMap[nodeID],
				SortOrder:       i,
				CreatedAt:       now,
			}
		}
		return tx.Create(&bindings).Error
	})
}

func (r *PortProfileRepositoryImpl) GetProfileNodeIDs(ctx context.Context, profileID int) ([]int, error) {
	nodes, err := r.GetProfileNodes(ctx, profileID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	return ids, nil
}

func (r *PortProfileRepositoryImpl) GetProfileNodes(ctx context.Context, profileID int) ([]*clash.ProxyNode, error) {
	var bindings []portProfileNodeModel
	if err := r.db.WithContext(ctx).
		Where("port_profile_id = ?", profileID).
		Order("sort_order asc, id asc").
		Find(&bindings).Error; err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		return []*clash.ProxyNode{}, nil
	}

	nodeIDs := make([]int, 0, len(bindings))
	fingerprints := make([]string, 0, len(bindings))
	for _, b := range bindings {
		if b.ProxyNodeID > 0 {
			nodeIDs = append(nodeIDs, b.ProxyNodeID)
		}
		if b.NodeFingerprint != "" {
			fingerprints = append(fingerprints, b.NodeFingerprint)
		}
	}

	var models []proxyNodeModel
	if err := r.db.WithContext(ctx).
		Where("id IN ? OR fingerprint IN ?", nodeIDs, fingerprints).
		Find(&models).Error; err != nil {
		return nil, err
	}

	byIDMap := make(map[int]*proxyNodeModel, len(models))
	byFPMap := make(map[string]*proxyNodeModel, len(models))
	for i := range models {
		byIDMap[models[i].ID] = &models[i]
		if models[i].Fingerprint != "" {
			byFPMap[models[i].Fingerprint] = &models[i]
		}
	}

	result := make([]*clash.ProxyNode, 0, len(bindings))
	seenNodes := make(map[int]struct{})

	for _, b := range bindings {
		var matched *proxyNodeModel
		if m, ok := byIDMap[b.ProxyNodeID]; ok {
			matched = m
		} else if b.NodeFingerprint != "" {
			if m, ok := byFPMap[b.NodeFingerprint]; ok {
				matched = m
			}
		}
		if matched != nil {
			if _, seen := seenNodes[matched.ID]; !seen {
				seenNodes[matched.ID] = struct{}{}
				result = append(result, toProxyNodeDomain(matched))
			}
		}
	}

	return result, nil
}

// PortProfileTemplateRepositoryImpl
type PortProfileTemplateRepositoryImpl struct {
	db *gorm.DB
}

func NewPortProfileTemplateRepository(db *Database) repository.PortProfileTemplateRepository {
	return &PortProfileTemplateRepositoryImpl{db: db.db}
}

func (r *PortProfileTemplateRepositoryImpl) Create(ctx context.Context, item *clash.PortProfileTemplate) error {
	m := toPortProfileTemplateModel(item)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	item.ID = m.ID
	item.CreatedAt = m.CreatedAt
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *PortProfileTemplateRepositoryImpl) GetByID(ctx context.Context, id int) (*clash.PortProfileTemplate, error) {
	var m portProfileTemplateModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return toPortProfileTemplateDomain(&m), nil
}

func (r *PortProfileTemplateRepositoryImpl) List(ctx context.Context) ([]*clash.PortProfileTemplate, error) {
	var models []portProfileTemplateModel
	if err := r.db.WithContext(ctx).Order("id desc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*clash.PortProfileTemplate, len(models))
	for i, m := range models {
		items[i] = toPortProfileTemplateDomain(&m)
	}
	return items, nil
}

func (r *PortProfileTemplateRepositoryImpl) Delete(ctx context.Context, id int) error {
	res := r.db.WithContext(ctx).Delete(&portProfileTemplateModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// RuntimeConfigRepositoryImpl
type RuntimeConfigRepositoryImpl struct {
	db *gorm.DB
}

func NewRuntimeConfigRepository(db *Database) repository.RuntimeConfigRepository {
	return &RuntimeConfigRepositoryImpl{db: db.db}
}

func (r *RuntimeConfigRepositoryImpl) Upsert(ctx context.Context, item *clash.RuntimeConfig) error {
	m := &runtimeConfigModel{
		PortProfileID:  item.PortProfileID,
		KernelType:     item.KernelType,
		Checksum:       item.Checksum,
		RenderedConfig: item.RenderedConfig,
		UpdatedAt:      time.Now(),
	}
	if item.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	} else {
		m.CreatedAt = item.CreatedAt
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "port_profile_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"kernel_type", "checksum", "rendered_config", "updated_at"}),
	}).Create(m).Error
	if err != nil {
		return err
	}
	item.ID = m.ID
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *RuntimeConfigRepositoryImpl) GetByPortProfileID(ctx context.Context, profileID int) (*clash.RuntimeConfig, error) {
	var m runtimeConfigModel
	if err := r.db.WithContext(ctx).First(&m, "port_profile_id = ?", profileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return toRuntimeConfigDomain(&m), nil
}

func (r *RuntimeConfigRepositoryImpl) ListAll(ctx context.Context) ([]*clash.RuntimeConfig, error) {
	var models []runtimeConfigModel
	if err := r.db.WithContext(ctx).Order("id desc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*clash.RuntimeConfig, len(models))
	for i, m := range models {
		items[i] = toRuntimeConfigDomain(&m)
	}
	return items, nil
}

func (r *RuntimeConfigRepositoryImpl) DeleteByPortProfileID(ctx context.Context, profileID int) error {
	return r.db.WithContext(ctx).Delete(&runtimeConfigModel{}, "port_profile_id = ?", profileID).Error
}

// KernelInstanceRepositoryImpl
type KernelInstanceRepositoryImpl struct {
	db *gorm.DB
}

func NewKernelInstanceRepository(db *Database) repository.KernelInstanceRepository {
	return &KernelInstanceRepositoryImpl{db: db.db}
}

func (r *KernelInstanceRepositoryImpl) Upsert(ctx context.Context, item *clash.KernelInstance) error {
	m := toKernelInstanceModel(item)
	m.UpdatedAt = time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "kernel_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "pid", "work_dir", "config_path", "controller_address", "controller_secret", "active_config_checksum", "active_profile_count", "active_listener_count", "last_action", "last_error", "last_started_at", "last_stopped_at", "last_reloaded_at", "updated_at"}),
	}).Create(m).Error
	if err != nil {
		return fmt.Errorf("upsert kernel instance: %w", err)
	}
	item.ID = m.ID
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *KernelInstanceRepositoryImpl) GetByType(ctx context.Context, kernelType string) (*clash.KernelInstance, error) {
	var m kernelInstanceModel
	if err := r.db.WithContext(ctx).First(&m, "kernel_type = ?", kernelType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return toKernelInstanceDomain(&m), nil
}
