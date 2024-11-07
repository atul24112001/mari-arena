package lib

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const AdminPublicKey = "CVdndsAGyNj8BvLhtrQBLMtrwEgy53ACXFQmQMfH2MFQ"

var Pool *pgxpool.Pool
var UnderMaintenance = false

func ConnectDB() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")
	log.Println(dsn)

	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	fmt.Println("Connected to PostgreSQL successfully!")
}
