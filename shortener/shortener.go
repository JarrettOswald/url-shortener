package shortener

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"time"
	"url-shortener/storage"
	
	"github.com/redis/go-redis/v9"
)

type Shortener struct {
	db *redis.Client
}

func New() *Shortener {
	cfg := storage.Config{
		Addr:        "localhost:6379",
		Password:    "test1234",
		User:        "default",
		DB:          0,
		MaxRetries:  5,
		DialTimeout: 10 * time.Second,
		Timeout:     5 * time.Second,
	}

	db, err := storage.NewClient(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	return &Shortener{db: db}
}

func (s *Shortener) Shorten(input string) string {
	algorithm := fnv.New64a()
	algorithm.Write([]byte(input))
	number := algorithm.Sum64()

	return toBase62(number)
}

func toBase62(n uint64) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const base = uint64(len(charset))

	if n == 0 {
		return string(charset[0])
	}

	var sb strings.Builder
	for n > 0 {
		rem := n % base
		sb.WriteByte(charset[rem])
		n = n / base
	}

	return sb.String()
}

type ShortenRequest struct {
	URL string `json:"url"`
}

func (s *Shortener) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreate(w, r)
	case http.MethodGet:
		s.handleRedirect(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Shortener) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "Field 'url' is required", http.StatusBadRequest)
		return
	}

	shortKey := s.Shorten(req.URL)

	if err := s.db.Set(r.Context(), shortKey, req.URL, 0).Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	fullURL := fmt.Sprintf("http://%s/%s", r.Host, shortKey)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
}

func (s *Shortener) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortKey := strings.TrimPrefix(r.URL.Path, "/")
	if shortKey == "" {
		http.Error(w, "Short key missing", http.StatusNotFound)
		return
	}

	originalURL, err := s.db.Get(r.Context(), shortKey).Result()
	if err == redis.Nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}
