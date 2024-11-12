package handle

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mwantia/valkey-snapshot/pkg/config"
)

func Handle(cfg *config.SnapshotServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(cfg); err != nil {
			http.Error(w, fmt.Sprintf("Failed to return running config: %v", err), http.StatusInternalServerError)
		}
	}
}
