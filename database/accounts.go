package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prizmsol/prizmsol-server/graph/model"
	"github.com/prizmsol/prizmsol-server/utils"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

type SignedUser struct {
	ID       string
	Email    string
	Username string
}

func (db *BUN) createStripeCustomer(user model.User) model.User {
	stripe_secret := os.Getenv("STRIPE_SECRET_KEY")
	stripe.Key = stripe_secret

	params := &stripe.CustomerParams{
		Email: &user.Email,
		Name:  &user.Name,
	}

	c, err := customer.New(params)

	if err != nil {
		fmt.Println("Could not create Customer: ", err)
	}

	// update user database entry.
	row, err := db.client.NewUpdate().Model(&user).Set("stripe_customer_id = ?", c.ID).Where("id = ?", user.ID).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not update stripe customer id in users table: ", err)
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong when fetching updated rows affected: ", err)
	}

	if rows > 0 {
		fmt.Print("Successfully created stripe customer id")
	}

	return user
}

func (db *BUN) InsertUsers(input *model.NewUser) *model.User {
	var now = time.Now()
	id := uuid.New()

	data := model.User{
		ID:        id.String(),
		Name:      input.Name,
		Email:     input.Email,
		Username:  input.Username,
		CreatedAt: &now,
	}

	query := "INSERT INTO Users (id, name, email, username, created_at) VALUES ($1, $2, $3, $4, $5)"

	stmt, err := db.client.Prepare(query)

	if err != nil {
		fmt.Print("Error: ", err)
		return &model.User{}
	}

	defer stmt.Close()

	if _, err := stmt.Exec(data.ID, data.Name, data.Email, data.Username, data.CreatedAt); err != nil {
		fmt.Print("Error Could not save users into the database: ", err)
		return &model.User{}
	}

	// create stripe customer.
	data = db.createStripeCustomer(data)

	return &data
}

func (db *BUN) GetAccounts(limit int) ([]*model.User, error) {
	var users []*model.User
	row := db.client.NewSelect().Model(&users).Limit(limit).Scan(context.Background())

	if row != nil {
		fmt.Print("\n Error found when querying all users: ", row)
		return []*model.User{}, row
	}
	return users, nil
}

func (db *BUN) GetAccount(username string) (*model.User, error) {
	var result model.User
	row := db.client.NewSelect().Model(&result).Where("username = ?", username).Scan(context.Background())

	if row != nil {
		fmt.Print("\n Error found when query for users using username: ", row)
		return &model.User{}, row
	}

	return &result, nil
}

func (db *BUN) GetAccountByEmail(email string) (*model.User, error) {
	var result model.User
	row := db.client.NewSelect().Model(&result).Where("email = ?", email).Scan(context.Background())

	if row != nil {
		fmt.Print("\n Error found when selecting users by email: ", row)
		return &model.User{}, row
	}

	return &result, nil
}

func (db *BUN) LoginAccount(email string) (string, error) {

	// load env
	envErr := godotenv.Load()
	now := time.Now()

	if envErr != nil {
		fmt.Print("dotenv couldn't load")
	}

	var user model.User

	rows, err := db.client.NewUpdate().Model(&user).Set("last_login = ?", now).Where("email = ?", email).Returning("*").Exec(context.Background())

	if err != nil {
		return "Error in updating login", err
	}

	row, err := rows.RowsAffected()

	if err != nil {
		fmt.Println("No rows updated: ", err)
		return "No User found", err
	}

	if row > 0 {
		// create token and return.
		token := utils.EncodeToString(6)

		data := model.Token{
			UserID:    user.ID,
			Token:     token,
			CreatedAt: &now,
		}

		row, err := db.client.NewInsert().Model(&data).Exec(context.Background())

		if err != nil {
			fmt.Println("Could not insert token into db")
			return "Error", nil
		}

		rows, _ := row.RowsAffected()

		if rows > 0 {

			// send email.
			utils.SendMail(
				user.Email,
				"PrizmSol Login Verification",
				"<h3>Your Login code is: </h3><br /><h1>"+token+"</h1>",
				"Your login code is: "+token,
			)

			return user.ID, nil
		}
	}

	return "No user found", nil
}

func (db *BUN) VerifyToken(id string, token string) (string, error) {
	var user model.User
	var tokenModel model.Token

	count, err := db.client.NewSelect().Model(&tokenModel).Where("token = ?", token).Where("user_id = ?", id).ScanAndCount(context.Background())

	if err != nil {
		fmt.Println("Could not fetch token: ", err)
		return "No Token", err
	}

	if count > 0 {
		count, err := db.client.NewSelect().Model(&user).Where("id = ?", id).ScanAndCount(context.Background())

		if err != nil {
			fmt.Println("Could not fetch user: ", err)
			return "No User", err
		}

		if count > 0 {
			expiry := time.Now().AddDate(1, 0, 0).Unix()
			// fetch user and compile JWT.
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"user": SignedUser{
					ID:       user.ID,
					Email:    user.Email,
					Username: user.Username,
				},
				"exp": expiry,
				"nbf": time.Now().Unix(),
			})

			jwt_secret := []byte(os.Getenv("JWT_SECRET"))
			tokenString, err := tok.SignedString(jwt_secret)

			if err != nil {
				fmt.Println("Error signing token")
				return "Error Signing token", err
			}

			// delete token
			db.client.NewDelete().Model(&tokenModel).Where("token = ?", token).Scan(context.Background())

			return tokenString, nil
		}
	} else {
		return "No Token found", nil
	}

	return "Failed", nil
}

func (db *BUN) VerifyEmail(id string, email string) (bool, error) {
	var user model.User

	row, err := db.client.NewUpdate().Model(&user).Set("is_email_verified = ?", true).Where("email = ?", email).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not verify email. Error: ", err)
		return false, nil
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Error fetching rows affected in verifyEmail: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) SearchUsers(query string) ([]*model.User, error) {
	var users []*model.User

	err := db.client.NewSelect().Model(&users).Where("name LIKE ? OR biography LIKE ? OR username LIKE ? OR email LIKE ?", query, query, query, query).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not search Users. An error occured. Error: ", err)
		return nil, err
	}

	return users, nil
}
