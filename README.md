# Valkey Snapshot

```yml
address: :8080
interval: 60m
endpoints:
  - name: valkey1
    endpoint: 127.0.0.1:6379
    password: ''
    database: 0
    batchsize: 1000
backend:
  type: minio
  endpoint: 'http://127.0.0.1:9000'
  bucket: 'valkey-snapshots'
  accessKey: 'minioadmin'
  secretKey: 'minioadmin'
timestamp_format: 2006-01-02_15-04-05
```