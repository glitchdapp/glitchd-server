package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

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
	fmt.Println("Established a successful db connection!")

	return &BUN{client: conn}
}

func (db *BUN) createUserTable() {
	_, err := db.client.NewCreateTable().Model(&model.User{}).Table("users").IfNotExists().Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create users table ", err)
	}
}

func (db *BUN) createPostTable() {
	_, err := db.client.NewCreateTable().Model(&model.Post{}).Table("posts").IfNotExists().Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create posts table ", err)
	}
}

func (db *BUN) createMembershipTable() {
	_, err := db.client.NewCreateTable().Model(&model.Membership{}).Table("memberships").IfNotExists().Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create memberships table ", err)
	}
}

func (db *BUN) createTokenTable() {
	_, err := db.client.NewCreateTable().Model(&model.Token{}).Table("tokens").IfNotExists().Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create tokens table ", err)
	}
}

func (db *BUN) createWaitlistTable() {
	_, err := db.client.NewCreateTable().Model(&model.Waitlist{}).Table("waitlist").IfNotExists().Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create waitlist table ", err)
	}
}

func (db *BUN) InitTables() *BUN {
	// generate tables in db
	db.createUserTable()
	db.createTokenTable()
	db.createUserTable()
	db.createMembershipTable()
	db.createPostTable()
	db.createWaitlistTable()

	fmt.Println("Initialized Tables")

	return &BUN{db.client}
}
