package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/mwantia/valkey-snapshot/pkg/config"
)

func CreateClient(cfg config.SnapshotEndpointConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Endpoint,
		Password: cfg.Password,
		DB:       cfg.Database,
	})
}

func ScanKeys(ctx context.Context, rdb *redis.Client, cursor uint64, cfg config.SnapshotEndpointConfig) ([]string, uint64, error) {
	keys, next, err := rdb.Scan(ctx, cursor, "*", cfg.BatchSize).Result()
	if err != nil {
		return nil, cursor, fmt.Errorf("failed to scan keys: %v", err)
	}

	return keys, next, nil
}

func GetSnapshotKey(ctx context.Context, rdb *redis.Client, key string) (config.SnapshotKey, error) {
	snapshotkey := config.SnapshotKey{
		Key:      key,
		Metadata: map[string]string{},
	}
	keytype, err := rdb.Type(ctx, key).Result()
	if err != nil {
		return snapshotkey, err
	}

	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		ttl = -1
	}

	snapshotkey.Type = keytype
	snapshotkey.TTL = int64(ttl.Seconds())

	switch keytype {
	case "string":
		value, err := rdb.Get(ctx, key).Result()
		if err == nil {
			snapshotkey.Value = value
		}
	case "hash":
		hash, err := rdb.HGetAll(ctx, key).Result()
		if err == nil {
			snapshotkey.Metadata = hash
			buf, _ := json.Marshal(hash)
			snapshotkey.Value = string(buf)
		}
	case "list":
		listData, err := rdb.LRange(ctx, key, 0, -1).Result()
		if err == nil {
			buf, _ := json.Marshal(listData)
			snapshotkey.Value = string(buf)
		}
	case "set":
		members, err := rdb.SMembers(ctx, key).Result()
		if err == nil {
			buf, _ := json.Marshal(members)
			snapshotkey.Value = string(buf)
		}
	case "zset":
		data, err := rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
		if err == nil {
			buf, _ := json.Marshal(data)
			snapshotkey.Value = string(buf)
		}
	}

	return snapshotkey, nil
}
