package common

import (
	"fmt"
	"net/http"
)

func HealthCheckerHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
