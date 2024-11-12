package config

import "time"

type Snapshot struct {
	Metadata SnapshotMetadata `json:"metadata"`
	Keys     []SnapshotKey    `json:"keys"`
}

type SnapshotMetadata struct {
	Name      string    `json:"name"`
	Endpoint  string    `json:"endpoint"`
	Database  int       `json:"database"`
	CreatedAt time.Time `json:"created_at"`
	Count     int       `json:"count"`
}

type SnapshotKey struct {
	Key      string            `json:"key"`
	Value    string            `json:"value"`
	Type     string            `json:"type"`
	TTL      int64             `json:"ttl"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
