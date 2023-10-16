package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/prizmsol/prizmsol-server/graph/model"
)

func (db *BUN) UpdateMembership(id string, input model.UpdateMembership) (bool, error) {
	var membership model.Membership

	row, err := db.client.NewUpdate().Model(&membership).Set("status = ?", input.Status).Set("tier = ?", input.Tier).Set("stripe_subscription_id = ?", input.StripeSubscriptionID).Set("stripe_checkout_session = ?", input.StripeCheckoutSession).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not update membership: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in update memebership: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetMembership(id string) (*model.Membership, error) {
	var membership model.Membership

	err := db.client.NewSelect().Model(&membership).Where("app_id = ?", id).Scan(context.Background())

	if err != nil {
		fmt.Println("Error found when fetching membership getmembership: ", err)
		return nil, err
	}

	return &membership, nil
}

func (db *BUN) GetMembershipBySession(sessionId string) (*model.Membership, error) {
	var membership model.Membership

	err := db.client.NewSelect().Model(&membership).Where("stripe_checkout_session = ?", sessionId).Scan(context.Background())

	if err != nil {
		fmt.Println("Error found when fetching membership getmembershipbysessionid: ", err)
		return nil, err
	}

	return &membership, nil
}

func (db *BUN) CreateMembership(input model.NewMembership) (*model.Membership, error) {
	var app model.App
	id := uuid.New()

	// get app data.
	r := db.client.NewSelect().Model(&app).Where("id = ?", input.AppID).Scan(context.Background())

	if r != nil {
		fmt.Println("Error found when fetching app in createMembership: ", r)
		return &model.Membership{}, nil
	}

	data := model.Membership{
		ID:                    id.String(),
		AppID:                 app.ID,
		Tier:                  input.Tier,
		Status:                "pending",
		StripeSubscriptionID:  input.StripeSubscriptionID,
		StripeCheckoutSession: input.StripeCheckoutSession,
	}

	res, err := db.client.NewInsert().Model(&data).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create membership. Error: ", err)
		return &model.Membership{}, err
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Error fetching rows after inserting membership into db: ", err)
	}

	if row > 0 {
		fmt.Println("Created membership successfully: ", data.ID)
		return &data, nil
	}
	return &model.Membership{}, nil
}
