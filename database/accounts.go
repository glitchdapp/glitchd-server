package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/glitchd/glitchd-server/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

type CustomClaim struct {
	ID    string
	Email string
	jwt.RegisteredClaims
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
				"glitchd Login Verification",
				"<h3>Your Login code is: </h3><br /><h1>"+token+"</h1>",
				"Your login code is: "+token,
			)

			return user.ID, nil
		}
	}

	return "No user found", nil
}

func JwtGenerate(ctx context.Context, userID string, email string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaim{
		ID:    userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(1, 0, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	jwt_secret := []byte(os.Getenv("JWT_SECRET"))
	tokenString, err := t.SignedString(jwt_secret)

	if err != nil {
		fmt.Println("Error signing token")
		return "Error Signing token", err
	}

	return tokenString, nil
}

func JwtValidate(ctx context.Context, token string) (*jwt.Token, error) {
	jwt_secret := []byte(os.Getenv("JWT_SECRET"))
	return jwt.ParseWithClaims(token, &CustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There's a problem with the signing method: %v", token.Header["alg"])
		}
		return jwt_secret, nil
	})
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
			jwtString, err := JwtGenerate(context.Background(), user.ID, user.Email)

			if err == nil {
				// delete token
				db.client.NewDelete().Model(&tokenModel).Where("token = ?", token).Scan(context.Background())
			}

			return jwtString, nil
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

func (db *BUN) IsWaitlisted(email string) (bool, error) {
	var waitlist model.Waitlist

	res, err := db.client.NewSelect().Model(&waitlist).Where("email = ?", email).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not fetch waitlist entry: ", err)
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows affected: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) AddUserToWaitlist(input model.NewWaitlist) (*model.Waitlist, error) {
	var waitlist model.Waitlist
	var now = time.Now()
	id := uuid.New()

	if ok, err := db.IsWaitlisted(input.Email); ok {
		fmt.Println("User already waitlisted")
		return nil, err
	}

	data := model.Waitlist{
		ID:        id.String(),
		Email:     input.Email,
		CanEnter:  false,
		CreatedAt: &now,
	}

	row, err := db.client.NewInsert().Model(&data).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not insert user into waitlist: ", err)
		return nil, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows affected: ", err)
		return nil, err
	}

	if rows > 0 {
		return &data, nil
	}

	return &waitlist, nil
}

func (db *BUN) UpdateWaitlistEntry(email string, canEnter bool) (*model.Waitlist, error) {
	var waitlist model.Waitlist

	row, err := db.client.NewUpdate().Model(&waitlist).Set("can_enter = ?", canEnter).Where("email = ?", email).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not update waitlist entry: ", err)
		return nil, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows affected: ", err)
		return nil, err
	}

	if rows > 0 {
		fmt.Println("Successfully updated waitlist entry")
		return &waitlist, nil
	}

	return &waitlist, nil
}
