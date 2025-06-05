package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	csav1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/csa"
	jobv1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/job"
	nodev1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/node"

	"github.com/job-distributor/internal/config"
	"github.com/job-distributor/internal/server"
)

func main() {
	cfg := config.Load()

	log.Printf("🚀 Starting Job Distributor on port %s", cfg.Port)
	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create our JD implementation
	jdServer := server.NewJobDistributorServer(cfg)

	// Register services
	jobv1.RegisterJobServiceServer(grpcServer, jdServer)
	nodev1.RegisterNodeServiceServer(grpcServer, jdServer)
	csav1.RegisterCSAServiceServer(grpcServer, jdServer)

	// Enable reflection for easier testing
	reflection.Register(grpcServer)

	// Start listening
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("🛑 Shutting down Job Distributor...")
		grpcServer.GracefulStop()
	}()

	log.Printf("✅ Job Distributor ready on port %s", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
