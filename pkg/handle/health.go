package handle

import (
	"fmt"
	"net/http"
)

func HandleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{ 'result': 'OK' }")
	}
}
