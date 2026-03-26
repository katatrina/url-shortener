package mock

//go:generate mockgen -destination mock_url_repository.go -package mock github.com/katatrina/url-shortener/internal/service URLRepository
//go:generate mockgen -destination mock_user_repository.go -package mock github.com/katatrina/url-shortener/internal/service UserRepository
//go:generate mockgen -destination mock_token_maker.go -package mock github.com/katatrina/url-shortener/internal/token TokenMaker
//go:generate mockgen -destination mock_url_cache.go -package mock github.com/katatrina/url-shortener/internal/service URLCacheRepository
//go:generate mockgen -destination mock_click_event_query_repository.go -package mock github.com/katatrina/url-shortener/internal/service ClickEventQueryRepository
//go:generate mockgen -destination mock_url_stats_query_repository.go -package mock github.com/katatrina/url-shortener/internal/service URLStatsQueryRepository
//go:generate mockgen -destination mock_click_collector.go -package mock github.com/katatrina/url-shortener/internal/service ClickCollector
