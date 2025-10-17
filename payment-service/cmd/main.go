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
	"github.com/p40pmn/golang-grpc/internal/payment"
	"github.com/p40pmn/golang-grpc/internal/server"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	ln, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

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

	paymentSvc := payment.NewService(ctx, db)
	paymentServer := server.NewServer(paymentSvc)

	gServer := grpc.NewServer()
	pb.RegisterPaymentServiceServer(gServer, paymentServer)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- gServer.Serve(ln)
	}()

	log.Println("Server started and listening on port: 50051")

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
