package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/glitchd/glitchd-server/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/muxinc/mux-go/v5"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/uptrace/bun"
)

type CustomClaim struct {
	ID string
	jwt.RegisteredClaims
}

func (db *BUN) initializeChannel(user_id string) (bool, error) {
	id := uuid.New().String()
	now := time.Now()

	streamkey, playbackIds, err := db.createUserStreamInfo()

	res, err := db.client.NewRaw(
		"INSERT INTO channels (id, user_id, title, notification, category, streamkey, playback_id, tags, is_branded, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (user_id) DO UPDATE SET title=EXCLUDED.title, notification=EXCLUDED.notification, tags=EXCLUDED.tags",
		id, user_id, "", "", "", streamkey, playbackIds[0].Id, "", false, now,
	).Exec(context.Background())

	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) createUserStreamInfo() (string, []muxgo.PlaybackId, error) {
	client := muxgo.NewAPIClient(
		muxgo.NewConfiguration(muxgo.WithBasicAuth(os.Getenv("MUX_ACCESS"), os.Getenv("MUX_SECRET"))))

	// generate livestream.
	car := muxgo.CreateAssetRequest{PlaybackPolicy: []muxgo.PlaybackPolicy{muxgo.PUBLIC}}
	csr := muxgo.CreateLiveStreamRequest{NewAssetSettings: car, LowLatency: true, PlaybackPolicy: []muxgo.PlaybackPolicy{muxgo.PUBLIC}}
	r, err := client.LiveStreamsApi.CreateLiveStream(csr)

	if err != nil {
		return "", nil, err
	}

	return r.Data.StreamKey, r.Data.PlaybackIds, nil
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

	username := strings.Split(email, "@")
	data := model.User{
		ID:        id,
		Email:     email,
		Username:  username[0],
		CreatedAt: now,
	}

	res, err := db.client.NewRaw(
		"INSERT INTO ? (id, email, username, created_at) VALUES (?, ?, ?, ?)",
		bun.Ident("users"), id, email, username[0], now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create user. Error: ", err)
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong while saving the user: ", err)
	}

	if row > 0 {
		return &data, true
	}

	return nil, false
}

func (db *BUN) userWithEmailExists(email string) (int, error) {
	var user model.User
	count, err := db.client.NewSelect().Model(&user).Where("email = ?", email).ScanAndCount(context.Background())

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *BUN) CreateUser(input *model.NewUser) (string, error) {
	var now = time.Now()

	hasEmail, _ := db.userWithEmailExists(input.Email)

	if hasEmail > 0 {
		return "", errors.New("Email taken")
	}

	id := uuid.New().String()
	_, err := db.GetChatIdentity(id)

	if err != nil {
		fmt.Println("Could not load chat identity something went wrong. ", err)
		// update chat identity.
		input := model.ChatIdentityInput{
			Color: "#FF0000",
			Badge: "",
		}
		db.UpdateChatIdentity(id, input)
	}

	data := model.User{
		ID:        id,
		Name:      input.Name,
		Email:     input.Email,
		Username:  input.Username,
		Dob:       input.Dob,
		CreatedAt: now,
	}

	res, err := db.client.NewRaw(
		"INSERT INTO users (id, name, email, username, dob, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, input.Name, input.Email, input.Username, input.Dob, now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create user. Error: ", err)
		return "", err
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Something went wrong while saving the user: ", err)
	}

	if row > 0 {

		isMade, err := db.initializeChannel(id)

		if isMade {
			// send email.
			utils.SendMail(
				input.Email,
				"Welcome To Glitchd",
				"<h1>Welcome to Glitchd, "+input.Name+"!</h1><h4>We are glad you could join us.</h4>",
				"Welcome to Glitchd, "+input.Name+"!",
			)

			result, _ := db.createLoginToken(&data)

			return result, nil
		}

		return "", err
	}

	return "", nil
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

func (db *BUN) DeleteUser(id string) (bool, error) {
	var user model.User

	row, err := db.client.NewDelete().Model(&user).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Error deleting user: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
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
		// update chat identity.
		input := model.ChatIdentityInput{
			Color: "#FF0000",
			Badge: "",
		}
		db.UpdateChatIdentity(result.ID, input)
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
		return "", err
	}

	// proceed with login
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
		err := db.client.NewRaw("select * from ? where id = ?", bun.Ident("users"), id).Scan(context.Background(), &user)

		if err != nil {
			fmt.Println("Could not fetch user: ", err)
			return "No User", err
		}

		if user.ID != "" {
			jwtString, err := JwtGenerate(context.Background(), user.ID, user.Email)

			if err == nil {
				// delete token
				db.client.NewDelete().Model(&tokenModel).Where("token = ?", token).Scan(context.Background())
			}

			// let initialize chat identity if it does not exist.
			_, errs := db.GetChatIdentity(user.ID)
			if errs != nil {
				fmt.Println("Verify Account: Could not load chat identity something went wrong. ", errs)
				// update chat identity.
				input := model.ChatIdentityInput{
					Color: "#FF0000",
					Badge: "",
				}
				db.UpdateChatIdentity(user.ID, input)
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
