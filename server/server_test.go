package main

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "sampleManager.com/go-sampleManager-grpc/proto"
)

func setupTestDB(t *testing.T) (sqlmock.Sqlmock, *SampleManagerServiceServer) {
	mockDB, mock, err := sqlmock.New()
	assert.Nil(t, err, "Failed to create mock DB: %v", err)
	defer mockDB.Close()

	dialect := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	gormDb, err := gorm.Open(dialect, &gorm.Config{})
	assert.Nil(t, err, "Failed to open GORM DB: %v", err)

	server := &SampleManagerServiceServer{DB: gormDb}

	mock.ExpectBegin()
	return mock, server
}

func TestAddSampleMapping_WithEmptyClmSegments_returnsInvalidArgumentError(t *testing.T) {
	mock, server := setupTestDB(t)

	request := &pb.SampleMapping{
		ClmSegments:  []string{},
		ItemId:       "item_id",
		SampleItemId: "sample_item_id",
	}

	res, err := server.AddSampleMapping(context.Background(), request)

	mock.ExpectRollback()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestAddSampleMapping_WithEmptyItemId_returnsInvalidArgumentError(t *testing.T) {
	mock, server := setupTestDB(t)

	request := &pb.SampleMapping{
		ClmSegments:  []string{},
		ItemId:       "",
		SampleItemId: "sample_item_id",
	}

	res, err := server.AddSampleMapping(context.Background(), request)

	mock.ExpectRollback()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestAddSampleMapping_WithEmptySampleItemId_returnsInvalidArgumentError(t *testing.T) {
	mock, server := setupTestDB(t)

	request := &pb.SampleMapping{
		ClmSegments:  []string{},
		ItemId:       "item_id",
		SampleItemId: "",
	}

	res, err := server.AddSampleMapping(context.Background(), request)

	mock.ExpectRollback()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestAddSampleMapping_withDuplicateEntry_returnsError(t *testing.T) {
	mock, server := setupTestDB(t)

	request := &pb.SampleMapping{
		ClmSegments:  []string{"segment1", "segment2", "segment3"},
		ItemId:       "item_id",
		SampleItemId: "sample_item_id",
	}

	mock.ExpectExec(`INSERT INTO "sample_mappings"`).
		WithArgs(pq.StringArray([]string{"segment1", "segment2", "segment3"}), "item_id", "sample_item_id").
		WillReturnError(errors.New("duplicate key value violates unique constraint"))
	mock.ExpectCommit()

	res, err := server.AddSampleMapping(context.Background(), request)

	assert.NotNil(t, res)
	assert.False(t, res.Success)
	assert.Equal(t, res.Message, "Failed to add mapping to the database")
	assert.NotNil(t, err)
}

func TestAddSampleMapping_success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.Nil(t, err, "Failed to create mock DB: %v", err)
	defer mockDB.Close()

	dialect := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})
	gormDb, err := gorm.Open(dialect, &gorm.Config{})
	assert.Nil(t, err, "Failed to open GORM DB: %v", err)

	server := &SampleManagerServiceServer{DB: gormDb}

	request := &pb.SampleMapping{
		ClmSegments:  []string{"segment1", "segment2", "segment3"},
		ItemId:       "item_id",
		SampleItemId: "sample_item_id",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "sample_mappings"`).
		WithArgs(pq.StringArray([]string{"segment1", "segment2", "segment3"}), "item_id", "sample_item_id").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	res, err := server.AddSampleMapping(context.Background(), request)

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestFetchSampleItemId_WithEmptyClmSegments_returnsInvalidArgumentError(t *testing.T) {
	mock, server := setupTestDB(t)

	request := &pb.FetchSampleItemIdRequest{
		ClmSegments: []string{},
		ItemId:      "item_id",
	}

	res, err := server.GetSampleItemId(context.Background(), request)

	mock.ExpectRollback()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestFetchSampleItemId_WithEmptyItemId_returnsInvalidArgumentError(t *testing.T) {
	mock, server := setupTestDB(t)

	request := &pb.FetchSampleItemIdRequest{
		ClmSegments: []string{"segment1"},
		ItemId:      "",
	}

	res, err := server.GetSampleItemId(context.Background(), request)

	mock.ExpectRollback()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestFetchSampleItemId_mappingNotFound_returnsNotFoundError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.Nil(t, err, "Failed to create mock DB: %v", err)
	defer mockDB.Close()

	dialect := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	gormDb, err := gorm.Open(dialect, &gorm.Config{})
	assert.Nil(t, err, "Failed to open GORM DB: %v", err)

	server := &SampleManagerServiceServer{DB: gormDb}

	request := &pb.FetchSampleItemIdRequest{
		ClmSegments: []string{"segment1", "segment2"},
		ItemId:      "item_id",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_mappings" WHERE item_id = $1 AND clm_segments && $2`)).
		WithArgs("item_id", pq.Array([]string{"segment1", "segment2"})).
		WillReturnError(gorm.ErrRecordNotFound)

	response, err := server.GetSampleItemId(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.NotFound, statusErr.Code(), "Expected NotFound error")
	assert.Contains(t, err.Error(), "Mapping not found")
}

func TestFetchSampleItemId_unknownDatabaseError_returnsInternalError(t *testing.T) {
	mock, server := setupTestDB(t)

	request := &pb.FetchSampleItemIdRequest{
		ClmSegments: []string{"segment1", "segment2"},
		ItemId:      "item_id",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_mappings" item_id = $1 AND clm_segments && $2`)).
		WithArgs("item_id", pq.Array([]string{"segment1", "segment2"})).
		WillReturnError(errors.New("database error"))

	response, err := server.GetSampleItemId(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.Internal, statusErr.Code(), "Expected Internal error")
	assert.Contains(t, err.Error(), "Failed to fetch mapping from database")
}

func TestFetchSampleItemId_success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.Nil(t, err, "Failed to create mock DB: %v", err)
	defer mockDB.Close()

	dialect := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})
	gormDb, err := gorm.Open(dialect, &gorm.Config{})
	assert.Nil(t, err, "Failed to open GORM DB: %v", err)

	server := &SampleManagerServiceServer{DB: gormDb}

	request := &pb.FetchSampleItemIdRequest{
		ClmSegments: []string{"segment1", "segment2"},
		ItemId:      "item_id",
	}

	rows := sqlmock.NewRows([]string{"clm_segments", "item_id", "sample_item_id"}).AddRow(pq.StringArray([]string{"segment1", "segment2"}), "item_id", "sample_item_id")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_mappings" WHERE item_id = $1 AND clm_segments && $2`)).
		WithArgs("item_id", pq.Array([]string{"segment1", "segment2"})).
		WillReturnRows(rows)

	response, err := server.GetSampleItemId(context.Background(), request)

	assert.NoError(t, err)
	expectedResponse := &pb.FetchSampleItemIdResponse{
		SampleItemId: "sample_item_id",
	}
	assert.Equal(t, expectedResponse, response)
}
