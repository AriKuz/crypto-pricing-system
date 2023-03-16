package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"syscall"

	zero "github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/crypto-pricing-service/pricing_service/fetcher"
	pb "github.com/crypto-pricing-service/pricing_service/grpc"
)

type server struct {
	pb.PricingServiceServer
	lis    net.Listener
	server *grpc.Server
	coins  sync.Map
}

// Stop initiates graceful stop for the gRPC server
func (s *server) Stop() {
	s.server.GracefulStop()
}

// StartGrpcServer start listening on the initiated port and block to receive new connections
func (s *server) StartGrpcServer() {
	zero.Info().Msgf("starting server on:%s", s.lis.Addr())
	if err := s.server.Serve(s.lis); err != nil {
		zero.Error().Msgf("server terminated with error: %v", err)
		//sending
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}
}

// NewGrpcServer assigns a gRPC server on the default listening address
func NewGrpcServer(serverListenAddress, apiKey string) (*server, error) {
	s := &server{}
	var err error

	zero.Info().Msgf("listening on port:%s", serverListenAddress)
	if s.lis, err = net.Listen("tcp", serverListenAddress); err != nil {
		return nil, err
	}

	go fetcher.CheckCoinMarket(&s.coins, apiKey)
	s.server = grpc.NewServer()
	pb.RegisterPricingServiceServer(s.server, s)
	return s, nil
}

//GetPrices gets requests from gRPC client and sends back the price for requested token
func (s *server) GetPrices(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	var response pb.Response
	response.Token = req.TokenName
	price, ok := s.coins.Load(req.TokenName)
	if !ok {
		zero.Error().Msgf("GetPrices failed to load key:%s", req.TokenName)
		return nil, fmt.Errorf("GetPrices failed to load key:%s", req.TokenName)
	}
	response.Price, ok = price.(float64)
	if !ok {
		zero.Error().Msgf("GetPrices failed to convert value to float64")
		return nil, fmt.Errorf("GetPrices failed to convert value to float64")
	}
	return &response, nil
}
