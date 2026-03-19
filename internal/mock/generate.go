package mock

//go:generate mockgen -destination mock_url_repository.go -package mock github.com/katatrina/url-shortener/internal/service URLRepository
//go:generate mockgen -destination mock_user_repository.go -package mock github.com/katatrina/url-shortener/internal/service UserRepository
//go:generate mockgen -destination mock_token_maker.go -package mock github.com/katatrina/url-shortener/internal/token TokenMaker
//go:generate mockgen -destination mock_url_cache.go -package mock github.com/katatrina/url-shortener/internal/service URLCacheRepository
