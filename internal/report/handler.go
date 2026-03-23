package report

// Handler holds service dependency for the report module
type Handler struct {
service *Service
}

func NewHandler(service *Service) *Handler {
return &Handler{service: service}
}
