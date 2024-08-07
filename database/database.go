package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var DB *BUN

type BUN struct {
	client *bun.DB
}

func Connect() *BUN {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	err = client.Ping()
	if err != nil {
		panic(err)
	}
	conn := bun.NewDB(client, pgdialect.New())
	fmt.Println("Established Database Connection Successfully!")

	return &BUN{client: conn}
}
