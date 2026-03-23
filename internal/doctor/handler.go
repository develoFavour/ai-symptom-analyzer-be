package doctor

// Handler holds service dependency for the doctor module
type Handler struct {
service *Service
}

func NewHandler(service *Service) *Handler {
return &Handler{service: service}
}
