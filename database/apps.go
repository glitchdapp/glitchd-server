package database

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prizmsol/prizmsol-server/graph/model"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/checkout/session"
)

func (db *BUN) NewApp(input *model.NewApp) (*model.App, error) {
	var user model.User
	id := uuid.New()

	// get user data.
	r := db.client.NewSelect().Model(&user).Where("id = ?", input.OwnerID).Scan(context.Background())

	if r != nil {
		fmt.Println("Error found when fetching user in CreateApp: ", r)
		return &model.App{}, r
	}

	data := model.App{
		ID:          id.String(),
		Name:        input.Name,
		Description: &input.Description,
		Vanity:      input.Vanity,
		OwnerID:     user.ID,
		Tier:        &input.Tier,
	}

	res, err := db.client.NewInsert().Model(&data).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create app. Error: ", err)
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Error fetching rows after inserting app into db: ", err)
	}

	if row > 0 {
		fmt.Println("Created app successfully: ", input.Name)

		err := godotenv.Load()

		if err != nil {
			fmt.Println("Could not fetch dotenv file")
			return nil, err
		}

		// create checkout session.
		stripe_secret := os.Getenv("STRIPE_SECRET_KEY")
		stripe.Key = stripe_secret

		params := &stripe.CheckoutSessionParams{
			Customer: &user.StripeCustomerID,
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				Items: []*stripe.CheckoutSessionSubscriptionDataItemsParams{
					{
						Plan:     stripe.String("price_1NsvzXKLDOtuUibteb9kKaCb"),
						Quantity: stripe.Int64(1),
					},
				},
			},
			Mode:       stripe.String("subscription"),
			SuccessURL: stripe.String("http://localhost:3000/billing?session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:  stripe.String("http://localhost:3000/billing/cancel"),
		}

		s, err := session.New(params)
		if err != nil {
			fmt.Println("Something went wrong when creating the session: ", err)
			return nil, err
		}

		input := model.NewMembership{
			AppID:                 data.ID,
			UserID:                user.ID,
			Tier:                  input.Tier,
			StripeSubscriptionID:  "",
			StripeCheckoutSession: s.ID,
		}
		db.CreateMembership(input)
		return &data, nil
	}

	return &model.App{}, nil
}

func (db *BUN) UpdateApp(id string, input model.UpdateApp) (bool, error) {
	data := model.App(input)

	res, err := db.client.NewUpdate().Model(&data).Where("id = ?", id).Exec(context.Background())

	if err != nil {
		fmt.Println("Error updating app: ", err)
		return false, err
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Error fetching rows after updating app: ", err)
		return false, err
	}

	if row > 0 {
		fmt.Println("Updated app successfully: ", input.Name)
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetApps(user_id string) ([]*model.App, error) {
	var apps []*model.App
	err := db.client.NewSelect().Model(&apps).Where("owner_id = ?", user_id).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not get apps for users", user_id)
		fmt.Println("Error: ", err)
		return nil, err
	}

	return apps, nil
}

func (db *BUN) GetApp(id string) (*model.App, error) {
	var app model.App
	err := db.client.NewSelect().Model(&app).Where("id = ?", id).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not fetch app id: ", id)
		fmt.Println("Error: ", err)
		return nil, err
	}

	return &app, nil
}

func (db *BUN) GetAppByVanity(vanity string) (*model.App, error) {
	var app model.App
	err := db.client.NewSelect().Model(&app).Where("vanity = ?", vanity).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not fetch app with vanity: ", vanity)
		fmt.Println("Error: ", err)
		return nil, err
	}

	return &app, nil
}

func (db *BUN) GetUserApp(user_id string, vanity string) (*model.App, error) {
	var app model.App
	err := db.client.NewSelect().Model(&app).Where("vanity = ?", vanity).Where("owner_id = ?", user_id).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not fetch app with that userid and vanity")
		fmt.Println("Error: ", err)
		return nil, err
	}

	return &app, nil
}

func (db *BUN) GetUserApps(user_id string) ([]*model.App, error) {
	var apps []*model.App
	err := db.client.NewSelect().Model(&apps).Where("owner_id = ?", user_id).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not fetch apps for user")
		fmt.Println("Error: ", err)
		return nil, err
	}

	return apps, nil
}

func (db *BUN) DeleteApp(id string) (*model.App, error) {
	var app model.App
	res, err := db.client.NewDelete().Model(&app).Where("id = ?", id).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not detete app. ", err)
		return nil, err
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch deleted app rows. ", err)
		return nil, err
	}

	if row > 0 {
		fmt.Println("Deleted app with id ", id)
		return &app, nil
	}
	fmt.Println("Something went wrong while deleting.")
	return nil, nil
}
