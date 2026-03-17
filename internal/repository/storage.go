package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Storage[E any] struct {
	db    *sqlx.DB
	table string
}

func NewStorage[E any](db *sqlx.DB, table string) *Storage[E] {
	return &Storage[E]{db: db, table: table}
}

func (s *Storage[E]) GetRowsBy(ctx context.Context, whereClause string, args ...any) ([]E, error) {
	var entities []E
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", s.table, whereClause)
	if err := s.db.SelectContext(ctx, &entities, query, args...); err != nil {
		return nil, fmt.Errorf("Storage.GetByID: query: %w", err)
	}
	return entities, nil
}

func (s *Storage[E]) GetByID(ctx context.Context, id int64) (*E, error) {
	var entity E
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", s.table)
	if err := s.db.GetContext(ctx, &entity, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("Storage.GetByID: query: %w", err)
	}
	//log.Println("GetByID sql res", &entity)
	return &entity, nil
}

func (s *Storage[E]) UpdateByID(ctx context.Context, id int64, setClause string, args ...any) error {
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id=%d", s.table, setClause, id)
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Storage.UpdateByID: sql exec: %w", err)
	}
	return nil
}

func (s *Storage[E]) Create(ctx context.Context, entity E) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	columns := []string{}
	//values := []any{}
	indexes := []string{}

	var column string

	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}
		column = t.Field(i).Tag.Get("db")
		if strings.Contains(column, "omitempty") {
			continue
		}
		if column == "" {
			column = t.Field(i).Name
		}
		columns = append(columns, column)
		//values = append(values, v.Field(i).Interface())
		indexes = append(indexes, ":"+column)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", s.table, strings.Join(columns, ","), strings.Join(indexes, ","))
	_, err := s.db.NamedExecContext(ctx, query, entity)
	if err != nil {
		return fmt.Errorf("Storage.Create: exec query: %w", err)
	}

	return nil
}

func (s *Storage[E]) Search(ctx context.Context, tsQuery string, limit int) ([]E, error) {
	var entities []E
	var query string
	var err error
	if limit > 0 {
		query = fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE search_vector_en @@ to_tsquery('english', $1) 
			OR search_vector_ru @@ to_tsquery('russian', $1) 
		ORDER BY GREATEST(
			ts_rank(search_vector_en, to_tsquery('english', $1)),
			ts_rank(search_vector_ru, to_tsquery('russian', $1))
		) DESC
		LIMIT $2
	`, s.table)
		err = s.db.SelectContext(ctx, &entities, query, tsQuery)
	} else {
		query = fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE search_vector_en @@ to_tsquery('english', $1) 
			OR search_vector_ru @@ to_tsquery('russian', $1) 
		ORDER BY GREATEST(
			ts_rank(search_vector_en, to_tsquery('english', $1)),
			ts_rank(search_vector_ru, to_tsquery('russian', $1))
		) DESC
		`, s.table)
		err = s.db.SelectContext(ctx, &entities, query, tsQuery)
	}

	if err != nil {
		return nil, fmt.Errorf("Storage.Search: exec query: %w", err)
	}

	return entities, nil
}
