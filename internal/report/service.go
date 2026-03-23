package report

// Service holds repository dependency for the report module
type Service struct {
repo Repository
}

func NewService(repo Repository) *Service {
return &Service{repo: repo}
}
