package server

import (
	"context"

	pb "github.com/p40pmn/golang-grpc/genproto/go"
	"github.com/p40pmn/golang-grpc/internal/order"
)

type Server struct {
	order *order.Service
	pb.UnimplementedOrderServiceServer
}

func NewServer(order *order.Service) *Server {
	return &Server{
		order: order,
	}
}

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	o, err := s.order.CreateOrder(ctx, &order.OrderReq{
		OwnerID:    req.UserId,
		ProductIDs: req.ProductIds,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderResponse{
		OrderId: o.ID,
		Total:   o.TotalAmount,
		Status:  o.Status,
	}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	o, err := s.order.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrderResponse{
		OrderId:  o.ID,
		UserId:   o.OwnerID,
		Total:    o.TotalAmount,
		Status:   o.Status,
		Products: newPBProductsFromProduct(o.Products),
	}, nil
}

func newPBProductsFromProduct(ps []*order.Product) []*pb.Product {
	pbProducts := make([]*pb.Product, 0, len(ps))
	for _, p := range ps {
		pbProducts = append(pbProducts, &pb.Product{
			Id:   p.ID,
			Name: p.DisplayName,
		})
	}
	return pbProducts
}
