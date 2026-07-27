-- +goose Up
ALTER TABLE port_profile_nodes ADD COLUMN node_fingerprint VARCHAR(255) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_port_profile_nodes_fp ON port_profile_nodes (node_fingerprint);

-- +goose Down
DROP INDEX IF EXISTS idx_port_profile_nodes_fp;
ALTER TABLE port_profile_nodes DROP COLUMN IF EXISTS node_fingerprint;
