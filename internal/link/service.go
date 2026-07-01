package link

type Service struct {
	linkRepo *Repository
}

func NewService(linkRepo *Repository) *Service {
	return &Service{linkRepo: linkRepo}
}
