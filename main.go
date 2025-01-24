package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	"time"
)

type RequestLog struct {
	Timestamp time.Time           `json:"timestamp"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Headers   map[string][]string `json:"headers"`
	Body      interface{}         `json:"body"`
	FormData  map[string][]string `json:"form_data,omitempty"`
	Status    int                 `json:"status"`
}

func customLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Parse form data first
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil && err != http.ErrNotMultipart {
			fmt.Printf("Form parse error: %v\n", err)
		}

		// Read body for JSON/raw content
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))

		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err != nil {
			jsonBody = string(body)
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		log := RequestLog{
			Timestamp: start,
			Method:    r.Method,
			Path:      r.URL.Path,
			Headers:   r.Header,
			Body:      jsonBody,
			FormData:  r.Form,
			Status:    ww.Status(),
		}

		logJSON, err := json.MarshalIndent(log, "", "    ")
		if err != nil {
			fmt.Printf("Error marshaling log: %v\n", err)
			return
		}
		fmt.Printf("\n%s\n", string(logJSON))
	})
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"msg": "ok"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(customLogger)
	r.Use(middleware.Recoverer)
	r.HandleFunc("/*", handleCallback)

	fmt.Printf("Server started on http://0.0.0.0:8123\n")
	
	if err := http.ListenAndServe("0.0.0.0:8123", r); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
