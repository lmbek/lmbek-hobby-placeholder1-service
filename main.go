package main

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)

var products = []string{"widget", "gadget", "gizmo", "doohickey", "thingamajig"}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("homepage visited", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Placeholder1 Service is running"))
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("health check", "method", r.Method, "remote", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Placeholder1 Service is Healthy"))
	})

	mux.HandleFunc("GET /api/products", func(w http.ResponseWriter, r *http.Request) {
		delay := time.Duration(rand.IntN(200)) * time.Millisecond
		time.Sleep(delay)
		slog.Info("products listed", "count", len(products), "delay_ms", delay.Milliseconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"products":%d}`, len(products))
	})

	mux.HandleFunc("POST /api/orders", func(w http.ResponseWriter, r *http.Request) {
		delay := time.Duration(50+rand.IntN(300)) * time.Millisecond
		time.Sleep(delay)
		if rand.IntN(100) < 15 {
			slog.Error("order processing failed", "error", "payment gateway timeout", "delay_ms", delay.Milliseconds())
			http.Error(w, `{"error":"payment gateway timeout"}`, http.StatusServiceUnavailable)
			return
		}
		orderID := rand.IntN(90000) + 10000
		product := products[rand.IntN(len(products))]
		slog.Info("order created", "order_id", orderID, "product", product, "delay_ms", delay.Milliseconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"order_id":%d,"product":"%s","status":"confirmed"}`, orderID, product)
	})

	mux.HandleFunc("GET /api/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		delay := time.Duration(30+rand.IntN(150)) * time.Millisecond
		time.Sleep(delay)
		results := rand.IntN(50)
		slog.Info("search executed", "query", query, "results", results, "delay_ms", delay.Milliseconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"query":"%s","results":%d}`, query, results)
	})

	handler := loggingMiddleware(mux)

	// Start background traffic simulator if enabled
	if os.Getenv("ENABLE_TRAFFIC_SIMULATOR") == "true" {
		go simulateTraffic()
		slog.Info("traffic simulator enabled")
	} else {
		slog.Info("traffic simulator disabled, set ENABLE_TRAFFIC_SIMULATOR=true to enable")
	}

	slog.Info("Placeholder1 Service starting", "port", 8082)
	if err := http.ListenAndServe(":8082", handler); err != nil {
		slog.Error("Error starting service", "error", err)
	}
}

func simulateTraffic() {
	time.Sleep(3 * time.Second) // wait for server to start
	client := &http.Client{Timeout: 5 * time.Second}
	base := "http://localhost:8082"

	endpoints := []struct {
		method string
		path   string
		weight int
	}{
		{"GET", "/", 10},
		{"GET", "/health", 20},
		{"GET", "/api/products", 30},
		{"POST", "/api/orders", 25},
		{"GET", "/api/search?q=widget", 15},
		{"GET", "/api/search?q=gadget", 10},
		{"GET", "/api/search?q=gizmo", 5},
	}

	totalWeight := 0
	for _, e := range endpoints {
		totalWeight += e.weight
	}

	slog.Info("traffic simulator started", "endpoints", len(endpoints))

	for {
		pick := rand.IntN(totalWeight)
		cumulative := 0
		for _, e := range endpoints {
			cumulative += e.weight
			if pick < cumulative {
				req, _ := http.NewRequest(e.method, base+e.path, nil)
				req.Header.Set("User-Agent", "TrafficSimulator/1.0")
				resp, err := client.Do(req)
				if err != nil {
					slog.Warn("simulator request failed", "path", e.path, "error", err.Error())
				} else {
					resp.Body.Close()
				}
				break
			}
		}
		// Random interval between 500ms and 3s
		time.Sleep(time.Duration(500+rand.IntN(2500)) * time.Millisecond)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start).String(),
			"remote", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
