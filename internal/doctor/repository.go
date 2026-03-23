package doctor

import "gorm.io/gorm"

// Repository defines the DB contract for the doctor module
type Repository interface {
// TODO: define doctor DB operations
}

type postgresRepository struct {
db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
return &postgresRepository{db: db}
}
