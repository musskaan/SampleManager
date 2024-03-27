package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sampleManager.com/go-sampleManager-grpc/model"
)

const (
	host     = "localhost"
	port     = 5433
	username = "postgres"
	password = "pgpswd"
	dbName   = "SampleSegmentMappingDB"
	sslMode  = "disable"
)

func Connection() *gorm.DB {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, username, password, dbName, sslMode)

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	log.Println("Connected to the database")

	err = db.AutoMigrate(&model.SampleMapping{})
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	return db
}
