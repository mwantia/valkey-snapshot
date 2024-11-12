package backend

import "github.com/mwantia/valkey-snapshot/pkg/config"

func CreateBackend(cfg *config.SnapshotBackendConfig) (*S3BackendClient, error) {
	return CreateS3Backend(cfg)
}
