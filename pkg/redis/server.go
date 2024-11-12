package redis

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

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

	go DoBeginGoroutine(client, cfg)

	http.HandleFunc("/health", handle.HandleHealth())

	log.Printf("Starting server on '%s'", cfg.Address)
	return http.ListenAndServe(cfg.Address, nil)
}

func DoBeginGoroutine(client *backend.S3BackendClient, cfg *config.SnapshotServerConfig) {
	interval, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to parse scrape interval '%s'", cfg.Interval))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		for _, endpoint := range cfg.Endpoints {
			go func(endpoint config.SnapshotEndpointConfig) {
				conn, err := net.Dial("tcp", endpoint.Endpoint)
				if err != nil {
					log.Printf("Failed to connect to Redis endpoint: %v", err)
				}
				defer conn.Close()

				_, err = conn.Write([]byte("PSYNC ? -1\r\n"))
				if err != nil {
					log.Printf("Failed to send PSYNC command: %v", err)
				}

				timestamp := time.Now().Format(cfg.TimestampFormat)
				name := fmt.Sprintf("%s/snapshot_%s.rdb", endpoint.Name, timestamp)

				fmt.Println("Receiving RDB snapshot...")
				const readTimeout = 5 * time.Second

				var buffer bytes.Buffer
				buf := make([]byte, 4096)

				for {
					// Set a read deadline; if no data is received within 5 seconds, the read will fail
					if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
						log.Fatalf("Failed to set read deadline: %v", err)
					}

					// Read from the connection
					n, err := conn.Read(buf)
					if err != nil {
						// If the error is a timeout, we assume the transfer is complete
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							fmt.Println("No data received for 5 seconds. Ending snapshot transfer.")
							break
						}
						log.Fatalf("Error receiving RDB snapshot: %v", err)
					}

					// Write received data to the file
					if _, err := buffer.Write(buf[:n]); err != nil {
						log.Fatalf("Error writing data to file: %v", err)
					}
				}

				fmt.Println("RDB snapshot received successfully.")
				fmt.Println("Uploading RDB snapshot to MinIO...")

				size := int64(buffer.Len())
				_, err = client.Minio.PutObject(context.Background(), client.Config.Bucket, name, &buffer, size, minio.PutObjectOptions{
					ContentType: "application/octet-stream",
				})
				if err != nil {
					log.Printf("Failed to upload RDB snapshot to MinIO: %v", err)
				}

				fmt.Println("RDB snapshot uploaded to MinIO successfully.")
			}(endpoint)
		}
		<-ticker.C
	}
}
