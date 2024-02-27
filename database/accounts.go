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
		fmt.Println("Successfully created stripe customer id")
	}

	return user
}

func (db *BUN) generateToken(user model.User) (model.User, error) {
	var now = time.Now()
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
		return model.User{}, nil
	}

	rows, _ := row.RowsAffected()

	if rows > 0 {

		// send email.
		utils.SendMail(
			user.Email,
			"Glitchd Login Verification",
			"<h2>Your login code is: </h2><br /><h1>"+token+"</h1>",
			"Your login code is: "+token,
		)

		return user, nil
	}

	return model.User{}, nil
}

func (db *BUN) checkEmailExists(email string) (bool, error) {
	var user model.User
	count, err := db.client.NewSelect().Model(&user).Where("email = ?", email).ScanAndCount(context.Background())

	if err != nil {
		fmt.Println("Something went wrong. Could not check if email exists.", err)
		return false, err
	}

	if count > 0 {
		fmt.Println("Email exists go on :).")
		return true, nil
	}

	return false, nil
}

func (db *BUN) CreateUser(input model.NewUser) (*model.User, error) {
	var now = time.Now()
	id := uuid.New()

	emailExists, err := db.checkEmailExists(input.Email)

	if err != nil {
		fmt.Println("Something went wrong when checking for email.")
		return nil, nil
	}

	if emailExists == true {
		fmt.Println("Email exists let's get outta here.")
		return &model.User{}, nil
	}

	data := model.User{
		ID:        id.String(),
		Name:      input.Name,
		Email:     input.Email,
		Username:  input.Username,
		CreatedAt: &now,
	}

	res, err := db.client.NewInsert().Model(&data).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not insert new user: ", err)
		return &model.User{}, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not insert new user. Something went wrong.")
		return nil, err
	}

	if rows > 0 {
		fmt.Println("Created New User.")

		return &data, nil
	}

	return &model.User{}, nil
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

func (db *BUN) GetAccountByID(id string) (*model.User, error) {
	var result model.User
	row := db.client.NewSelect().Model(&result).Where("id = ?", id).Scan(context.Background())

	if row != nil {
		fmt.Print("\n Error found when selecting users by id: ", row)
		return &model.User{}, row
	}

	return &result, nil
}

func (db *BUN) LoginAccount(email string) (string, error) {

	// load env
	envErr := godotenv.Load()

	if envErr != nil {
		fmt.Print("dotenv couldn't load")
	}

	var user model.User

	count, err := db.client.NewSelect().Model(&user).Where("email = ?", email).ScanAndCount(context.Background())

	if err != nil {
		return "User does not exist", err
	}

	if count > 0 {
		newUser, err := db.generateToken(user)

		if err != nil {
			fmt.Println("Could not generate token")
		}

		return newUser.ID, nil
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
