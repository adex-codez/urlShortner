package server

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"urlShortner/cmd/web"
	"urlShortner/internal/database"

	"github.com/a-h/templ"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/shorten-url", s.GenerateShortUrlHandler)
	mux.HandleFunc("/", s.CatchAllHandler)

	fileServer := http.FileServer(http.FS(web.Files))
	mux.Handle("/assets/", fileServer)
	mux.Handle("/index", templ.Handler(web.UrlShortnerForm()))

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

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid Form data", http.StatusBadRequest)
	}

	url := r.FormValue("url")

	urlRegex := `https?://(?:www\.)?[a-zA-Z0-9-]+(\.[a-zA-Z]{2,})+(:[0-9]{1,5})?(/[^\s]*)?`
	re := regexp.MustCompile(urlRegex)
	isMatch := re.MatchString(url)

	if !isMatch {
		http.Error(w, "Invalid Url", http.StatusBadRequest)
		return
	}

	unique_code := randomString(4)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	res, err := s.db.Exec("insert into url (long_url, unique_code) values (?, ?)", url, unique_code)
	_ = res
	if err != nil {
		http.Error(w, "Unable to add to database", http.StatusInternalServerError)
		log.Printf(err.Error())
	}

	component := web.ShowUrl(fmt.Sprint(os.Getenv("URL"), unique_code))
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *Server) CatchAllHandler(w http.ResponseWriter, r *http.Request) {
	unique_code := strings.TrimPrefix(r.URL.Path, "/")
	if unique_code == "" || strings.Contains(unique_code, "/") {
		http.Error(w, "Invalid unique_code", http.StatusBadRequest)
		return
	}

	var url database.Url
	err := s.db.Get(&url, "Select long_url, unique_code from url WHERE unique_code=$1", unique_code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching url: %s", err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, url.LongUrl, http.StatusPermanentRedirect)
}
