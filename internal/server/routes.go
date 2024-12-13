package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/shorten-url", s.GenerateShortUrlHandler)

	mux.HandleFunc("/health", s.healthHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}

func (s *Server) GenerateShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type Url struct {
		Url string `json:"url"`
	}

	var url Url

	err := json.NewDecoder(r.Body).Decode(&url)

	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	urlRegex := `https?://(?:www\.)?[a-zA-Z0-9-]+(\.[a-zA-Z]{2,})+(:[0-9]{1,5})?(/[^\s]*)?`
	re := regexp.MustCompile(urlRegex)
	isMatch := re.MatchString(url.Url)

	if !isMatch {
		http.Error(w, "Invalid Url", http.StatusBadRequest)
		return
	}

	uniquecode := randomString(4)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	insertUrl = `insert into url (long_url, unique_code) values (?, ?)`

}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
