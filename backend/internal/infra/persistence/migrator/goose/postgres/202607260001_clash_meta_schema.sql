-- +goose Up
-- PostgreSQL Goose Migration for Clash Meta (Mihomo) features

CREATE TABLE IF NOT EXISTS source_configs (
    id BIGSERIAL PRIMARY KEY,
    source_type VARCHAR(32) NOT NULL DEFAULT 'upload',
    source_url VARCHAR(2048) NOT NULL DEFAULT '',
    content_type VARCHAR(255) NOT NULL DEFAULT '',
    fetched_at TIMESTAMPTZ,
    filename VARCHAR(255) NOT NULL DEFAULT '',
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    raw_content TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'parsed',
    total_nodes INTEGER NOT NULL DEFAULT 0,
    valid_nodes INTEGER NOT NULL DEFAULT 0,
    invalid_nodes INTEGER NOT NULL DEFAULT 0,
    duplicate_nodes INTEGER NOT NULL DEFAULT 0,
    imported_nodes INTEGER NOT NULL DEFAULT 0,
    uploaded_by VARCHAR(64) NOT NULL DEFAULT '',
    uploaded_by_id INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_source_configs_hash ON source_configs (content_hash);
CREATE INDEX IF NOT EXISTS idx_source_configs_status ON source_configs (status);
CREATE INDEX IF NOT EXISTS idx_source_configs_created ON source_configs (created_at DESC);

CREATE TABLE IF NOT EXISTS proxy_nodes (
    id BIGSERIAL PRIMARY KEY,
    source_config_id INTEGER NOT NULL DEFAULT 0,
    source_config_name VARCHAR(255) NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL DEFAULT '',
    type VARCHAR(64) NOT NULL DEFAULT '',
    server VARCHAR(255) NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    tags VARCHAR(255) NOT NULL DEFAULT '',
    fingerprint VARCHAR(64) NOT NULL DEFAULT '',
    metadata_json TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_test_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    last_latency_ms INTEGER,
    last_test_error TEXT NOT NULL DEFAULT '',
    last_tested_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_proxy_nodes_fingerprint ON proxy_nodes (fingerprint);
CREATE INDEX IF NOT EXISTS idx_proxy_nodes_source ON proxy_nodes (source_config_id);
CREATE INDEX IF NOT EXISTS idx_proxy_nodes_enabled ON proxy_nodes (enabled);
CREATE INDEX IF NOT EXISTS idx_proxy_nodes_test_status ON proxy_nodes (last_test_status);

CREATE TABLE IF NOT EXISTS node_test_results (
    id BIGSERIAL PRIMARY KEY,
    node_id INTEGER NOT NULL,
    test_type VARCHAR(32) NOT NULL DEFAULT 'delay',
    success BOOLEAN NOT NULL DEFAULT FALSE,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    tested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_node_test_results_node ON node_test_results (node_id, tested_at DESC);

CREATE TABLE IF NOT EXISTS port_profiles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL DEFAULT '',
    listen_host VARCHAR(120) NOT NULL DEFAULT '0.0.0.0',
    mixed_port INTEGER NOT NULL DEFAULT 0,
    socks_port INTEGER NOT NULL DEFAULT 0,
    http_port INTEGER NOT NULL DEFAULT 0,
    proxy_settings_json TEXT NOT NULL DEFAULT '',
    include_in_runtime BOOLEAN NOT NULL DEFAULT TRUE,
    kernel_type VARCHAR(32) NOT NULL DEFAULT 'mihomo',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_port_profiles_kernel ON port_profiles (kernel_type);

CREATE TABLE IF NOT EXISTS port_profile_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL DEFAULT '',
    description VARCHAR(255) NOT NULL DEFAULT '',
    mixed_port INTEGER NOT NULL DEFAULT 0,
    proxy_settings_json TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS port_profile_nodes (
    id BIGSERIAL PRIMARY KEY,
    port_profile_id INTEGER NOT NULL,
    proxy_node_id INTEGER NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_port_profile_nodes_profile ON port_profile_nodes (port_profile_id, sort_order ASC);
CREATE INDEX IF NOT EXISTS idx_port_profile_nodes_node ON port_profile_nodes (proxy_node_id);

CREATE TABLE IF NOT EXISTS runtime_configs (
    id BIGSERIAL PRIMARY KEY,
    port_profile_id INTEGER NOT NULL,
    kernel_type VARCHAR(32) NOT NULL DEFAULT 'mihomo',
    checksum VARCHAR(64) NOT NULL DEFAULT '',
    rendered_config TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_runtime_configs_profile ON runtime_configs (port_profile_id);

CREATE TABLE IF NOT EXISTS kernel_instances (
    id BIGSERIAL PRIMARY KEY,
    kernel_type VARCHAR(32) NOT NULL DEFAULT 'mihomo',
    status VARCHAR(32) NOT NULL DEFAULT 'stopped',
    pid INTEGER,
    work_dir VARCHAR(255) NOT NULL DEFAULT '',
    config_path VARCHAR(255) NOT NULL DEFAULT '',
    controller_address VARCHAR(255) NOT NULL DEFAULT '',
    controller_secret VARCHAR(255) NOT NULL DEFAULT '',
    active_config_checksum VARCHAR(64) NOT NULL DEFAULT '',
    active_profile_count INTEGER NOT NULL DEFAULT 0,
    active_listener_count INTEGER NOT NULL DEFAULT 0,
    last_action VARCHAR(32) NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    last_started_at TIMESTAMPTZ,
    last_stopped_at TIMESTAMPTZ,
    last_reloaded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_kernel_instances_type ON kernel_instances (kernel_type);

-- +goose Down
DROP TABLE IF EXISTS kernel_instances;
DROP TABLE IF EXISTS runtime_configs;
DROP TABLE IF EXISTS port_profile_nodes;
DROP TABLE IF EXISTS port_profile_templates;
DROP TABLE IF EXISTS port_profiles;
DROP TABLE IF EXISTS node_test_results;
DROP TABLE IF EXISTS proxy_nodes;
DROP TABLE IF EXISTS source_configs;
