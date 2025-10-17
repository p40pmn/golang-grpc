package order

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	pb "github.com/p40pmn/golang-grpc/genproto/go"
)

var ErrOrderNotFound = fmt.Errorf("order not found")

type Service struct {
	product pb.ProductServiceClient
	payment pb.PaymentServiceClient
	db      *pgxpool.Pool
}

func NewService(_ context.Context, db *pgxpool.Pool, product pb.ProductServiceClient, payment pb.PaymentServiceClient) *Service {
	return &Service{
		db:      db,
		product: product,
		payment: payment,
	}
}

func (s *Service) CreateOrder(ctx context.Context, in *OrderReq) (*Order, error) {
	o := newOrder(in)

	products, err := s.product.ListProducts(ctx, &pb.ListProductsRequest{
		Ids: in.ProductIDs,
	})
	if err != nil {
		return nil, err
	}

	o.SetTotalAmount(products.Products)

	payment, err := s.payment.Charge(ctx, &pb.ChargeRequest{
		OrderId: o.ID,
		Amount:  o.TotalAmount,
	})
	if err != nil {
		return nil, err
	}

	if payment.Success {
		o.Completed()
	}

	if err := createOrder(ctx, s.db, o); err != nil {
		return nil, err
	}

	return o, nil
}

func (s *Service) GetOrderByID(ctx context.Context, id string) (*Order, error) {
	o, err := getOrderByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	productResp, err := s.product.ListProducts(ctx, &pb.ListProductsRequest{Ids: o.ProductIDs})
	if err != nil {
		return nil, err
	}

	products := newProductsFromPBProducts(productResp.Products)
	o.SetProducts(products)
	return o, nil
}

type OrderReq struct {
	OwnerID    string   `json:"userId"`
	ProductIDs []string `json:"productIds"`
}

func newOrder(o *OrderReq) *Order {
	return &Order{
		ID:          genID(),
		Status:      "PENDING",
		OwnerID:     o.OwnerID,
		ProductIDs:  o.ProductIDs,
		TotalAmount: 0,
	}
}

func (o *Order) SetTotalAmount(ps []*pb.Product) {
	o.TotalAmount = sumAmountFromProducts(ps)
}

func (o *Order) Completed() {
	o.Status = "COMPLETED"
}

func (o *Order) SetProducts(ps []*Product) {
	o.Products = ps
}

func newProductFromPBFromPBProduct(p *pb.Product) *Product {
	return &Product{
		ID:          p.Id,
		DisplayName: p.Name,
		Price:       p.Price,
	}
}

func newProductsFromPBProducts(ps []*pb.Product) []*Product {
	products := make([]*Product, 0, len(ps))
	for _, p := range ps {
		products = append(products, newProductFromPBFromPBProduct(p))
	}

	return products
}

func sumAmountFromProducts(ps []*pb.Product) int32 {
	var sum int32
	for _, p := range ps {
		sum += p.Price
	}

	return sum
}

type Order struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	OwnerID     string   `json:"userId"`
	ProductIDs  []string `json:"productIds,omitempty"`
	TotalAmount int32    `json:"totalAmount"`

	Products []*Product `json:"products,omitempty"`
}

type Product struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Price       int32  `json:"price"`
}

func createOrder(ctx context.Context, db *pgxpool.Pool, o *Order) error {
	q, args := sq.Insert(`"order"`).
		Columns(
			"id",
			"owner_id",
			"product_ids",
			"total_amount",
			"status",
		).
		Values(
			o.ID,
			o.OwnerID,
			pq.Array(o.ProductIDs),
			o.TotalAmount,
			o.Status,
		).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	_, err := db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}

func getOrderByID(ctx context.Context, db *pgxpool.Pool, id string) (*Order, error) {
	q, args := sq.Select(
		"id",
		"owner_id",
		"product_ids",
		"total_amount",
		"status",
	).
		From(`"order"`).
		Where(sq.Eq{
			"id": id,
		}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		MustSql()

	var o Order
	err := db.QueryRow(ctx, q, args...).Scan(
		&o.ID,
		&o.OwnerID,
		&o.ProductIDs,
		pq.Array(&o.TotalAmount),
		&o.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &o, nil
}
