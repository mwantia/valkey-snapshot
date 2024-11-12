package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/minio/minio-go/v7"
	"github.com/mwantia/valkey-snapshot/pkg/backend"
	"github.com/mwantia/valkey-snapshot/pkg/config"
	"github.com/mwantia/valkey-snapshot/pkg/handle"
)

func Start(cfg *config.SnapshotServerConfig) error {
	client, err := backend.CreateBackend(cfg.Backend)
	if err != nil {
		return err
	}

	go BeginGoroutine(client, cfg)

	http.HandleFunc("/health", handle.Health())
	http.Handle("/", handle.Handle(cfg))

	log.Printf("Starting server on '%s'", cfg.Address)
	return http.ListenAndServe(cfg.Address, nil)
}

func BeginGoroutine(client *backend.S3BackendClient, cfg *config.SnapshotServerConfig) {
	interval, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		log.Fatalf("unable to parse scrape interval '%s'", cfg.Interval)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := RunGoroutine(client, cfg); err != nil {
			log.Fatalf("unable to continue goroutine: %v", err)
		}
		<-ticker.C
	}
}

func RunGoroutine(client *backend.S3BackendClient, cfg *config.SnapshotServerConfig) error {
	for _, endpoint := range cfg.Endpoints {
		go func(endpoint config.SnapshotEndpointConfig) {
			rdb := CreateClient(endpoint)
			ctx := context.Background()

			fmt.Printf("Creating snapshot for '%s'\n", endpoint.Name)

			snapshot, err := CreateSnapshot(ctx, rdb, endpoint)
			if err != nil {
				fmt.Printf("Unable to create snapshot: %v\n", err)
			}

			fmt.Println("Snapshot created successfully")
			fmt.Println("Uploading snapshot to backend")

			name := fmt.Sprintf("%s/%v/snapshot_%s.json", endpoint.Name, endpoint.Database, snapshot.Metadata.CreatedAt.Format(cfg.TimestampFormat))
			if err := UploadSnapshot(ctx, *client, name, snapshot); err != nil {
				fmt.Printf("Unable to upload snapshot: %v\n", err)
			}

			fmt.Println("Snapshot uploaded successfully")
		}(endpoint)
	}
	return nil
}

func CreateSnapshot(ctx context.Context, rdb *redis.Client, cfg config.SnapshotEndpointConfig) (config.Snapshot, error) {
	snapshot := config.Snapshot{
		Metadata: config.SnapshotMetadata{
			Name:      cfg.Name,
			Endpoint:  cfg.Endpoint,
			Database:  cfg.Database,
			CreatedAt: time.Now(),
		},
		Keys: make([]config.SnapshotKey, 0),
	}

	var cursor uint64
	for {
		keys, next, err := ScanKeys(ctx, rdb, cursor, cfg)
		if err != nil {
			return snapshot, err
		}

		for _, key := range keys {
			snapshotKey, err := GetSnapshotKey(ctx, rdb, key)
			if err != nil {
				continue
			}

			snapshot.Keys = append(snapshot.Keys, snapshotKey)
			snapshot.Metadata.Count++
		}

		if next == 0 {
			break
		}
		cursor = next
	}

	return snapshot, nil
}

func UploadSnapshot(ctx context.Context, client backend.S3BackendClient, name string, snapshot config.Snapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	size := int64(len(data))
	reader := bytes.NewReader(data)

	_, err = client.Minio.PutObject(ctx, client.Config.Bucket, name, reader, size, minio.PutObjectOptions{
		ContentType: "application/json",
	})
	return err
}
