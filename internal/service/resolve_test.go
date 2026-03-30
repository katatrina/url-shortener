package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katatrina/url-shortener/internal/cache"
	"github.com/katatrina/url-shortener/internal/model"
	mock "github.com/katatrina/url-shortener/internal/service/mock"
	"go.uber.org/mock/gomock"
)

func newTestService(t *testing.T) (
	*Service,
	*mock.MockURLRepository,
	*mock.MockURLCacheRepository,
	*mock.MockClickCollector,
) {
	ctrl := gomock.NewController(t)

	urlRepo := mock.NewMockURLRepository(ctrl)
	urlCache := mock.NewMockURLCacheRepository(ctrl)
	collector := mock.NewMockClickCollector(ctrl)

	svc := New(urlRepo, nil, urlCache, nil, nil, nil, collector)
	return svc, urlRepo, urlCache, collector
}

var testMeta = model.ClickMeta{
	IP:        "127.0.0.1",
	UserAgent: "Mozilla/5.0",
	Referer:   "https://example.com",
}

func TestResolve_CacheHit(t *testing.T) {
	svc, _, urlCache, collector := newTestService(t)

	urlCache.EXPECT().
		Get(gomock.Any(), "abc1234").
		Return(&cache.CachedURL{
			ID:          "url-id-1",
			OriginalURL: "https://example.com",
		}, nil)

	collector.EXPECT().
		Track("url-id-1", testMeta)

	got, err := svc.Resolve(context.Background(), "abc1234", testMeta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com" {
		t.Fatalf("got %q, want %q", got, "https://example.com")
	}
}

func TestResolve_CacheHitExpired(t *testing.T) {
	svc, _, urlCache, _ := newTestService(t)

	past := time.Now().Add(-1 * time.Hour).Unix()
	urlCache.EXPECT().
		Get(gomock.Any(), "abc1234").
		Return(&cache.CachedURL{
			ID:          "url-id-1",
			OriginalURL: "https://example.com",
			ExpiresAt:   &past,
		}, nil)

	_, err := svc.Resolve(context.Background(), "abc1234", testMeta)
	if !errors.Is(err, model.ErrURLExpired) {
		t.Fatalf("got %v, want ErrURLExpired", err)
	}
}

func TestResolve_CacheMissThenDB(t *testing.T) {
	svc, urlRepo, urlCache, collector := newTestService(t)

	urlCache.EXPECT().
		Get(gomock.Any(), "abc1234").
		Return(nil, nil) // cache miss

	urlRepo.EXPECT().
		FindByShortCode(gomock.Any(), "abc1234").
		Return(&model.URL{
			ID:          "url-id-1",
			ShortCode:   "abc1234",
			OriginalURL: "https://example.com",
		}, nil)

	urlCache.EXPECT().
		Set(gomock.Any(), "abc1234", gomock.Any()).
		Return(nil)

	collector.EXPECT().
		Track("url-id-1", testMeta)

	got, err := svc.Resolve(context.Background(), "abc1234", testMeta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com" {
		t.Fatalf("got %q, want %q", got, "https://example.com")
	}
}

func TestResolve_CacheErrorFallsBackToDB(t *testing.T) {
	svc, urlRepo, urlCache, collector := newTestService(t)

	urlCache.EXPECT().
		Get(gomock.Any(), "abc1234").
		Return(nil, errors.New("redis connection refused"))

	urlRepo.EXPECT().
		FindByShortCode(gomock.Any(), "abc1234").
		Return(&model.URL{
			ID:          "url-id-1",
			ShortCode:   "abc1234",
			OriginalURL: "https://example.com",
		}, nil)

	urlCache.EXPECT().
		Set(gomock.Any(), "abc1234", gomock.Any()).
		Return(nil)

	collector.EXPECT().
		Track("url-id-1", testMeta)

	got, err := svc.Resolve(context.Background(), "abc1234", testMeta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com" {
		t.Fatalf("got %q, want %q", got, "https://example.com")
	}
}

func TestResolve_URLNotFound(t *testing.T) {
	svc, urlRepo, urlCache, _ := newTestService(t)

	urlCache.EXPECT().
		Get(gomock.Any(), "noexist").
		Return(nil, nil)

	urlRepo.EXPECT().
		FindByShortCode(gomock.Any(), "noexist").
		Return(nil, model.ErrURLNotFound)

	_, err := svc.Resolve(context.Background(), "noexist", testMeta)
	if !errors.Is(err, model.ErrURLNotFound) {
		t.Fatalf("got %v, want ErrURLNotFound", err)
	}
}

func TestResolve_DBURLExpired(t *testing.T) {
	svc, urlRepo, urlCache, _ := newTestService(t)

	past := time.Now().Add(-1 * time.Hour)
	urlCache.EXPECT().
		Get(gomock.Any(), "abc1234").
		Return(nil, nil)

	urlRepo.EXPECT().
		FindByShortCode(gomock.Any(), "abc1234").
		Return(&model.URL{
			ID:          "url-id-1",
			ShortCode:   "abc1234",
			OriginalURL: "https://example.com",
			ExpiresAt:   &past,
		}, nil)

	_, err := svc.Resolve(context.Background(), "abc1234", testMeta)
	if !errors.Is(err, model.ErrURLExpired) {
		t.Fatalf("got %v, want ErrURLExpired", err)
	}
}

func TestResolve_NilCache(t *testing.T) {
	ctrl := gomock.NewController(t)

	urlRepo := mock.NewMockURLRepository(ctrl)
	collector := mock.NewMockClickCollector(ctrl)
	svc := New(urlRepo, nil, nil, nil, nil, nil, collector) // nil cache

	urlRepo.EXPECT().
		FindByShortCode(gomock.Any(), "abc1234").
		Return(&model.URL{
			ID:          "url-id-1",
			ShortCode:   "abc1234",
			OriginalURL: "https://example.com",
		}, nil)

	collector.EXPECT().
		Track("url-id-1", testMeta)

	got, err := svc.Resolve(context.Background(), "abc1234", testMeta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com" {
		t.Fatalf("got %q, want %q", got, "https://example.com")
	}
}
