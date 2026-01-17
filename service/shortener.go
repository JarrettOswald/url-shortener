package service

import (
	"context"
	"hash/fnv"
	"strings"
	"url-shortener/repository"
)

type ShortenerService interface {
	ShortenURL(ctx context.Context, originalURL string) string
	GetOriginalURL(ctx context.Context, shortKey string) (string, error)
}

type service struct {
	repo repository.Repository
}

func NewShortenerService(repo repository.Repository) ShortenerService {
	return &service{repo: repo}
}

func (s *service) ShortenURL(ctx context.Context, originalURL string) string {
	shortKey := s.generateKey(originalURL)

	s.repo.Save(ctx, shortKey, originalURL, 0)
	return shortKey
}

func (s *service) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	return s.repo.Get(ctx, shortKey)
}

func (s *service) generateKey(input string) string {
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
