package main

import (
	"context"
	"errors"
	"log"
	"net"

	"github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	database "sampleManager.com/go-sampleManager-grpc/db"
	"sampleManager.com/go-sampleManager-grpc/model"
	pb "sampleManager.com/go-sampleManager-grpc/proto"
)

type SampleManagerServiceServer struct {
	DB *gorm.DB
	pb.SampleManagerServiceServer
}

func (s *SampleManagerServiceServer) AddSampleMapping(ctx context.Context, request *pb.SampleMapping) (*pb.AddSampleMappingResponse, error) {
	clmSegments := request.ClmSegments
	itemId := request.ItemId
	sampleItemId := request.SampleItemId

	if len(clmSegments) == 0 || itemId == "" || sampleItemId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid arguments: %v, %v, %v", clmSegments, itemId, sampleItemId)
	}

	newMapping := &model.SampleMapping{
		ClmSegments:  clmSegments,
		ItemId:       itemId,
		SampleItemId: sampleItemId,
	}

	if err := s.DB.Create(newMapping).Error; err != nil {
		return &pb.AddSampleMappingResponse{
			Success: false,
			Message: "Failed to add mapping to the database",
		}, err
	}

	return &pb.AddSampleMappingResponse{
		Success: true,
		Message: "Mapping added successfully",
	}, nil
}

func (s *SampleManagerServiceServer) GetSampleItemId(ctx context.Context, request *pb.FetchSampleItemIdRequest) (*pb.FetchSampleItemIdResponse, error) {
	clmSegments := request.ClmSegments
	itemId := request.ItemId

	if len(clmSegments) == 0 || itemId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid arguments: %v, %v", clmSegments, itemId)
	}

	var mapping model.SampleMapping
	if err := s.DB.Where("item_id = ? AND clm_segments && ?", itemId, pq.StringArray(clmSegments)).Find(&mapping).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Mapping not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to fetch mapping from database: %v", err)
	}

	return &pb.FetchSampleItemIdResponse{
		SampleItemId: mapping.SampleItemId,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8002")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	db := database.Connection()

	pb.RegisterSampleManagerServiceServer(grpcServer, &SampleManagerServiceServer{DB: db})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve 8002: %v", err)
	}
}
