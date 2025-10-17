package product

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProductNotFound = fmt.Errorf("product not found")

type Service struct {
	db *pgxpool.Pool
}

func NewService(_ context.Context, db *pgxpool.Pool) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	p.ID = genID()
	err := createProduct(ctx, s.db, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

type ProductQuery struct {
	IDs []string
}

func (q *ProductQuery) ToSql() (string, any, error) {
	and := sq.And{}
	if len(q.IDs) > 0 {
		and = append(and, sq.Eq{"id": q.IDs})
	}

	return and.ToSql()
}

func (s *Service) ListProducts(ctx context.Context, in *ProductQuery) ([]*Product, error) {
	return listProducts(ctx, s.db, in)
}

func (s *Service) GetProduct(ctx context.Context, id string) (*Product, error) {
	return getProductByID(ctx, s.db, id)
}

type Product struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Price       int32  `json:"price"`
}

func createProduct(ctx context.Context, db *pgxpool.Pool, p *Product) error {
	q, args := sq.Insert("product").
		Columns(
			"id",
			"display_name",
			"price",
		).
		Values(
			p.ID,
			p.DisplayName,
			p.Price,
		).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	_, err := db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}

func listProducts(ctx context.Context, db *pgxpool.Pool, in *ProductQuery) ([]*Product, error) {
	pred, args, err := in.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	q, args := sq.Select(
		"id",
		"display_name",
		"price",
	).
		From("product").
		Where(pred, args).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	ps := make([]*Product, 0)
	rows, err := db.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID,
			&p.DisplayName,
			&p.Price,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		ps = append(ps, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return ps, nil
}

func getProductByID(ctx context.Context, db *pgxpool.Pool, id string) (*Product, error) {
	q, args := sq.Select(
		"id",
		"display_name",
		"price",
	).
		From("product").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	var p Product
	err := db.QueryRow(ctx, q, args...).Scan(
		&p.ID,
		&p.DisplayName,
		&p.Price,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &p, nil
}
