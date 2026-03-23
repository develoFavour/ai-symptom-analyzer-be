package diagnosis

import "gorm.io/gorm"

// Repository defines the DB contract for the diagnosis module
type Repository interface {
// TODO: define diagnosis DB operations
}

type postgresRepository struct {
db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
return &postgresRepository{db: db}
}
