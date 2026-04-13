package category

type Service interface {
	GetCategories(limit, offset int) ([]*Category, error)
	GetCategoryBreakdown(userID string, limit, offset int) ([]*CategoryBreakdown, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetCategories(limit, offset int) ([]*Category, error) {
	return s.repo.GetAll(limit, offset)
}

func (s *service) GetCategoryBreakdown(userID string, limit, offset int) ([]*CategoryBreakdown, error) {
	return s.repo.GetBreakdownByUserID(userID, limit, offset)
}
