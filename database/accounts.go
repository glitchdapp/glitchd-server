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
	"github.com/uptrace/bun"
)

type CustomClaim struct {
	ID string
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
		fmt.Println("Successfully created stripe customer id")
	}

	return user
}

func (db *BUN) registerUser(email string) (*model.User, bool) {
	var now = time.Now()
	id := uuid.New().String()

	data := model.User{
		ID:        id,
		Email:     email,
		Username:  id,
		CreatedAt: now,
	}

	res, err := db.client.NewRaw(
		"INSERT INTO ? (id, email, username, created_at) VALUES (?, ?, ?, ?)",
		bun.Ident("users"), id, email, id, now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create user. Error: ", err)
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong while saving the user: ", err)
	}

	if row > 0 {
		fmt.Println("Created User successfully")
		return &data, true
	}

	return nil, false
}

func (db *BUN) CreateUsers(input *model.NewUser) (*model.User, error) {
	var now = time.Now()

	data := model.User{
		ID:        uuid.New().String(),
		Email:     input.Email,
		CreatedAt: now,
	}

	res, err := db.client.NewInsert().Model(&data).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create user. Error: ", err)
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong while saving the user: ", err)
	}

	if row > 0 {
		fmt.Println("Created User successfully")
		return &data, nil
	}

	return &data, nil
}

func (db *BUN) UpdateUser(id string, input *model.UpdateUser) (bool, error) {
	var now = time.Now()

	data := model.User{}

	row, err := db.client.NewUpdate().Model(&data).Where("id = ?", id).Set("name = ?", input.Name).Set("username = ?", input.Username).Set("biography = ?", input.Biography).Set("updated_at = ?", now).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Error updating user: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong. ", err)
		return false, err
	}

	if rows > 0 {
		fmt.Println("Updated User Successfully...")
		return true, nil
	}

	return true, nil
}

func (db *BUN) UpdateUserStripe(id string, input *model.UserStripeInput) (bool, error) {
	var now = time.Now()

	row, err := db.client.NewRaw(
		"UPDATE users SET stripe_customer_id = ?, stripe_connected_link = ?, updated_at = ? WHERE id = ?",
		input.StripeCustomerID, input.StripeConnectedLink, now, id,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Error updating user: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong. ", err)
		return false, err
	}

	if rows > 0 {
		fmt.Println("Updated User Successfully...", input.StripeConnectedLink)
		return true, nil
	}

	return true, nil
}

func (db *BUN) UpdateUserPhoto(id string, photo string) (bool, error) {
	var now = time.Now()

	data := model.User{}

	row, err := db.client.NewUpdate().Model(&data).Where("id = ?", id).Set("photo = ?", photo).Set("updated_at = ?", now).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Error updating user photo: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong. ", err)
		return false, err
	}

	if rows > 0 {
		fmt.Println("Updated User Photo Successfully...")
		return true, nil
	}

	return true, nil
}

func (db *BUN) UpdateUserCoverPhoto(id string, photo string) (bool, error) {
	var now = time.Now()

	data := model.User{}

	row, err := db.client.NewUpdate().Model(&data).Where("id = ?", id).Set("cover = ?", photo).Set("updated_at = ?", now).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Error updating user cover photo: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong. ", err)
		return false, err
	}

	if rows > 0 {
		fmt.Println("Updated User Cover Photo Successfully...")
		return true, nil
	}

	return true, nil
}

func (db *BUN) DeleteUser(id string) (*model.User, error) {
	var user model.User

	row, err := db.client.NewDelete().Model(&user).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Error deleting user: ", err)
		return nil, err
	}

	rows, err := row.RowsAffected()

	if rows > 0 {
		return &user, nil
	}

	return &model.User{}, nil
}

func (db *BUN) GetAccounts(limit int) ([]*model.User, error) {
	var users []*model.User
	row := db.client.NewRaw("SELECT * FROM ? LIMIT ?", bun.Ident("users"), limit).Scan(context.Background(), &users)

	if row != nil {
		fmt.Print("\n Error found when querying all users: ", row)
		return []*model.User{}, row
	}
	return users, nil
}

func (db *BUN) GetUser(id string) (*model.User, error) {
	var result model.User

	err := db.client.NewRaw("SELECT * FROM users WHERE id = ?", id).Scan(context.Background(), &result)

	if err != nil {
		fmt.Print("\n Error found when querying for users with id: ", err)
		return nil, err
	}

	chat_identity, err := db.GetChatIdentity(result.ID)
	if err != nil {
		fmt.Println("Could not load chat identity something went wrong. ", err)
	}

	result.ChatIdentity = chat_identity

	return &result, nil
}

func (db *BUN) GetAccount(email string) (*model.User, error) {
	var result model.User
	row := db.client.NewRaw("SELECT * FROM users WHERE email = ?", email).Scan(context.Background(), &result)

	if row != nil {
		fmt.Print("\n Error found when query for users using email: ", row)
		return nil, row
	}

	chat_identity, err := db.GetChatIdentity(result.ID)
	if err != nil {
		fmt.Println("Could not load chat identity something went wrong. ", err)
	}

	result.ChatIdentity = chat_identity

	return &result, nil
}

func (db *BUN) GetAccountByEmail(email string) (*model.User, error) {
	var result model.User
	row := db.client.NewRaw("SELECT * FROM users WHERE email = ?", email).Scan(context.Background(), &result)

	if row != nil {
		fmt.Print("\n Error found when selecting users by email: ", row)
		return nil, row
	}

	chat_identity, err := db.GetChatIdentity(result.ID)
	if err != nil {
		fmt.Println("Could not load chat identity something went wrong. ", err)
	}

	result.ChatIdentity = chat_identity

	return &result, nil
}

func (db *BUN) GetUserByUsername(username string) (*model.User, error) {
	var result model.User
	row := db.client.NewRaw("SELECT * FROM users WHERE lower(username) = lower(?)", username).Scan(context.Background(), &result)
	if row != nil {
		fmt.Print("\n Error found when selecting users by username: ", row)
		return nil, row
	}

	chat_identity, err := db.GetChatIdentity(result.ID)
	if err != nil {
		fmt.Println("Could not load chat identity something went wrong. ", err)
	}

	result.ChatIdentity = chat_identity

	return &result, nil
}

func (db *BUN) createLoginToken(user *model.User) (string, error) {

	now := time.Now()
	// create token and return.
	token := utils.EncodeToString(6)

	data := model.Token{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     token,
		CreatedAt: &now,
	}

	row, err := db.client.NewInsert().Model(&data).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not insert token into db. ", err)
		return "Login Error: Could not insert token", nil
	}

	rows, _ := row.RowsAffected()

	if rows > 0 {

		// send email.
		utils.SendMail(
			user.Email,
			"Glitchd Login Verification",
			"<h3>Your Login code is: </h3><br /><h1>"+token+"</h1>",
			"Your login code is: "+token,
		)

		return user.ID, nil
	}

	return "Something went wrong", nil
}

func (db *BUN) LoginAccount(email string) (string, error) {

	// load env
	envErr := godotenv.Load()

	if envErr != nil {
		fmt.Print("dotenv couldn't load")
	}

	var user model.User

	err := db.client.NewRaw("SELECT * FROM ? WHERE email = ?", bun.Ident("users"), email).Scan(context.Background(), &user)

	if err != nil {
		if user.ID != "" {
			fmt.Println("No user present...")
			fmt.Println("Inserting new user...")
			user, isRegistered := db.registerUser(email)
			if isRegistered {
				fmt.Println("Registered user successfully...")
			}

			result, err := db.createLoginToken(user)

			return result, err
		}
		return "", err
	}

	result, err := db.createLoginToken(&user)

	return result, err
}

func JwtGenerate(ctx context.Context, userID string, email string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaim{
		ID: userID,
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
			return nil, fmt.Errorf("there's a problem with the signing method: %v", token.Header["alg"])
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

	err := db.client.NewRaw("SELECT * FROM users WHERE LOWER(name) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)", "%"+query+"%", "%"+query+"%", query).Scan(context.Background(), &users)

	if err != nil {
		fmt.Println("Could not search Users. An error occured. Error: ", err)
		return nil, err
	}

	return users, nil
}
