package doctor

// Service holds repository dependency for the doctor module
type Service struct {
repo Repository
}

func NewService(repo Repository) *Service {
return &Service{repo: repo}
}
