package server

import (
	"context"

	pb "github.com/p40pmn/golang-grpc/genproto/go"
	"github.com/p40pmn/golang-grpc/internal/payment"
)

type Server struct {
	payment payment.Service
	pb.UnimplementedPaymentServiceServer
}

func NewServer(payment *payment.Service) *Server {
	return &Server{
		payment: *payment,
	}
}

func (s *Server) Charge(ctx context.Context, req *pb.ChargeRequest) (*pb.ChargeResponse, error) {
	p, err := s.payment.CreatePayment(ctx, &payment.Payment{
		OrderID: req.OrderId,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &pb.ChargeResponse{Success: true, TransactionId: p.ID}, nil
}
