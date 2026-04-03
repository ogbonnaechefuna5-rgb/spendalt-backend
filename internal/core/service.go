package core

// Service provides base operations over a core Repository.
// Domain services embed this to inherit GetByID and Delete.
type Service[T any] struct {
	Repo *Repository[T]
}

func (s *Service[T]) GetByID(id int) (*T, error) {
	return s.Repo.GetByID(id)
}

func (s *Service[T]) Delete(id int) error {
	return s.Repo.Delete(id)
}
