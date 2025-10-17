package payment

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Payment struct {
	ID      string `json:"id"`
	OrderID string `json:"orderId"`
	Amount  int32  `json:"amount"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(_ context.Context, db *pgxpool.Pool) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) CreatePayment(ctx context.Context, p *Payment) (*Payment, error) {
	p.ID = genID()
	err := createPayment(ctx, s.db, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func createPayment(ctx context.Context, db *pgxpool.Pool, p *Payment) error {
	q, args := sq.Insert("payment").
		Columns(
			"id",
			"order_id",
			"amount",
		).
		Values(
			p.ID,
			p.OrderID,
			p.Amount,
		).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	_, err := db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}
