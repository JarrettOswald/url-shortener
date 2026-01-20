package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockRepository - мок репозитория для тестирования
type MockRepository struct {
	SaveFunc func(ctx context.Context, key string, url string, ttl time.Duration) error
	GetFunc  func(ctx context.Context, key string) (string, error)
}

func (m *MockRepository) Save(ctx context.Context, key string, url string, ttl time.Duration) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, key, url, ttl)
	}
	return nil
}

func (m *MockRepository) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", errors.New("not found")
}

func TestShortenURL(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
		saveError   error
		wantError   bool
	}{
		{
			name:        "successful URL shortening",
			originalURL: "https://example.com",
			saveError:   nil,
			wantError:   false,
		},
		{
			name:        "another URL",
			originalURL: "https://github.com/user/repo",
			saveError:   nil,
			wantError:   false,
		},
		{
			name:        "same URL produces same key",
			originalURL: "https://example.com",
			saveError:   nil,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{
				SaveFunc: func(ctx context.Context, key string, url string, ttl time.Duration) error {
					return tt.saveError
				},
			}

			service := NewShortenerService(mockRepo)
			ctx := context.Background()

			shortKey := service.ShortenURL(ctx, tt.originalURL)

			if shortKey == "" {
				t.Error("expected non-empty short key")
			}

			// Проверяем, что ключ содержит только base62 символы
			const base62Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, char := range shortKey {
				found := false
				for _, validChar := range base62Charset {
					if char == validChar {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("short key contains invalid character: %c", char)
				}
			}
		})
	}
}

func TestShortenURL_ConsistentHashing(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewShortenerService(mockRepo)
	ctx := context.Background()

	url := "https://example.com"
	key1 := service.ShortenURL(ctx, url)
	key2 := service.ShortenURL(ctx, url)

	if key1 != key2 {
		t.Errorf("expected same key for same URL, got %s and %s", key1, key2)
	}
}

func TestShortenURL_DifferentURLs(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewShortenerService(mockRepo)
	ctx := context.Background()

	url1 := "https://example.com"
	url2 := "https://example.org"

	key1 := service.ShortenURL(ctx, url1)
	key2 := service.ShortenURL(ctx, url2)

	if key1 == key2 {
		t.Error("expected different keys for different URLs")
	}
}

func TestGetOriginalURL(t *testing.T) {
	tests := []struct {
		name        string
		shortKey    string
		mockReturn  string
		mockError   error
		expectedURL string
		expectedErr bool
	}{
		{
			name:        "successful retrieval",
			shortKey:    "abc123",
			mockReturn:  "https://example.com",
			mockError:   nil,
			expectedURL: "https://example.com",
			expectedErr: false,
		},
		{
			name:        "key not found",
			shortKey:    "notfound",
			mockReturn:  "",
			mockError:   errors.New("redis: nil"),
			expectedURL: "",
			expectedErr: true,
		},
		{
			name:        "repository error",
			shortKey:    "error",
			mockReturn:  "",
			mockError:   errors.New("connection error"),
			expectedURL: "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{
				GetFunc: func(ctx context.Context, key string) (string, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			service := NewShortenerService(mockRepo)
			ctx := context.Background()

			url, err := service.GetOriginalURL(ctx, tt.shortKey)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if url != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
			}
		})
	}
}

func TestToBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "a",
		},
		{
			name:     "small number",
			input:    10,
			expected: "k",
		},
		{
			name:     "large number",
			input:    123456789,
			expected: "HUawi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBase62(tt.input)
			if result != tt.expected {
				t.Errorf("toBase62(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	mockRepo := &MockRepository{}
	s := &service{repo: mockRepo}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple URL",
			input: "https://example.com",
		},
		{
			name:  "long URL",
			input: "https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := s.generateKey(tt.input)

			if key == "" {
				t.Error("expected non-empty key")
			}

			// Проверяем идемпотентность
			key2 := s.generateKey(tt.input)
			if key != key2 {
				t.Error("generateKey should produce consistent results")
			}
		})
	}
}

func TestNewShortenerService(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewShortenerService(mockRepo)

	if service == nil {
		t.Error("expected non-nil service")
	}
}

func BenchmarkShortenURL(b *testing.B) {
	mockRepo := &MockRepository{}
	service := NewShortenerService(mockRepo)
	ctx := context.Background()
	url := "https://example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ShortenURL(ctx, url)
	}
}

func BenchmarkToBase62(b *testing.B) {
	var num uint64 = 123456789

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toBase62(num)
	}
}
