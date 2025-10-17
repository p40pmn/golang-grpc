package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/p40pmn/golang-grpc/genproto/go"
	"github.com/p40pmn/golang-grpc/internal/order"
	"github.com/p40pmn/golang-grpc/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	dbCfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to parse database url: %v", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Connected to DB successfully!")

	productConn, err := grpc.NewClient(os.Getenv("PRODUCT_URL"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial grpc connection with product: %v", err)
	}
	defer productConn.Close()

	paymentConn, err := grpc.NewClient(os.Getenv("PAYMENT_URL"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial grpc connection with payment: %v", err)
	}
	defer paymentConn.Close()

	productClient := pb.NewProductServiceClient(productConn)
	paymentClient := pb.NewPaymentServiceClient(paymentConn)

	ln, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	orderSvc := order.NewService(ctx, db, productClient, paymentClient)

	gServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(gServer, server.NewServer(orderSvc))
	if err := gServer.Serve(ln); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- gServer.Serve(ln)
	}()

	log.Println("Server started and listening on port: 50055")

	select {
	case <-ctx.Done():
		log.Println("Waiting for server to shut down...")
		gServer.GracefulStop()
		log.Println("Server shut down")

	case err := <-errCh:
		if err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("failed to run server: %v", err)
		}
	}
}
