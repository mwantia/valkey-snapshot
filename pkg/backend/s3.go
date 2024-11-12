package backend

import (
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mwantia/valkey-snapshot/pkg/config"
)

type S3BackendClient struct {
	Minio  *minio.Client
	Config *config.SnapshotBackendConfig
}

func CreateS3Backend(cfg *config.SnapshotBackendConfig) (*S3BackendClient, error) {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	endpoint := u.Hostname()
	if u.Port() != "" {
		endpoint = u.Hostname() + ":" + u.Port()
	}

	m, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, cfg.Region),
		Secure: u.Scheme == "https",
	})
	if err != nil {
		return nil, err
	}

	return &S3BackendClient{
		Config: cfg,
		Minio:  m,
	}, nil
}
