package main

import (
	"log/slog"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User Service is Healthy"))
	})

	slog.Info("User Service starting", "port", 8082)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		slog.Error("Error starting service", "error", err)
	}
}
