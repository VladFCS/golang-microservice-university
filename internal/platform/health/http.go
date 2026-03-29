package health

import (
	"encoding/json"
	"net/http"
)

const Path = "/healthz"

type Response struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func Handler(service string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(Path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		WriteResponse(w, service)
	})

	return mux
}

func WriteResponse(w http.ResponseWriter, service string) {
	body, err := json.Marshal(Response{
		Status:  "ok",
		Service: service,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(body, '\n'))
}
