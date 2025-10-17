package server

import (
	"context"

	pb "github.com/p40pmn/golang-grpc/genproto/go"
	"github.com/p40pmn/golang-grpc/internal/product"
)

type Server struct {
	product *product.Service
	pb.UnimplementedProductServiceServer
}

func NewServer(product *product.Service) *Server {
	return &Server{
		product: product,
	}
}

func (s *Server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, err := s.product.ListProducts(ctx, &product.ProductQuery{IDs: req.Ids})
	if err != nil {
		return nil, err
	}

	return &pb.ListProductsResponse{
		Products: newPBProducts(products),
	}, nil
}

func (s *Server) GetProductByID(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := s.product.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetProductResponse{
		Product: newPBProduct(p),
	}, nil
}

func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	p, err := s.product.CreateProduct(ctx, toProduct(req.Product))
	if err != nil {
		return nil, err
	}
	return &pb.CreateProductResponse{Product: newPBProduct(p)}, nil
}

func toProduct(p *pb.Product) *product.Product {
	return &product.Product{
		ID:          p.Id,
		DisplayName: p.Name,
		Price:       p.Price,
	}
}

func newPBProducts(products []*product.Product) []*pb.Product {
	ps := make([]*pb.Product, 0, len(products))

	for _, p := range products {
		ps = append(ps, newPBProduct(p))
	}

	return ps
}

func newPBProduct(p *product.Product) *pb.Product {
	return &pb.Product{
		Id:    p.ID,
		Name:  p.DisplayName,
		Price: p.Price,
	}
}
