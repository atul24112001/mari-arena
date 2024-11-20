package lib

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const AdminPublicKey = "CVdndsAGyNj8BvLhtrQBLMtrwEgy53ACXFQmQMfH2MFQ"

var Pool *pgxpool.Pool
var UnderMaintenance = false

func ConnectDB() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Connected to PostgreSQL successfully!")
}
