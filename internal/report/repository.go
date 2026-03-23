package report

import "gorm.io/gorm"

// Repository defines the DB contract for the report module
type Repository interface {
// TODO: define report DB operations
}

type postgresRepository struct {
db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
return &postgresRepository{db: db}
}
