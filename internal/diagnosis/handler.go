package diagnosis

// Handler holds service dependency for the diagnosis module
type Handler struct {
service *Service
}

func NewHandler(service *Service) *Handler {
return &Handler{service: service}
}
