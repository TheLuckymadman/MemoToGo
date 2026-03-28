package handler

import "context"

type Repository[E any] interface {
	GetRowsBy(ctx context.Context, whereClause string, args ...any) ([]E, error)
	GetByID(ctx context.Context, id int64) (*E, error)
	UpdateByID(ctx context.Context, id int64, setClause string, args ...any) error
	Create(ctx context.Context, entity E) error
	Search(ctx context.Context, tsQuery string, limit int) ([]E, error)
}
