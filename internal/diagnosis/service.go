package diagnosis

// Service holds repository dependency for the diagnosis module
type Service struct {
repo Repository
}

func NewService(repo Repository) *Service {
return &Service{repo: repo}
}
