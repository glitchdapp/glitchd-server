package database

import (
	"context"
	"fmt"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (db *BUN) CreatePayment(input model.PaymentInput) (bool, error) {

	id := uuid.New().String()
	now := time.Now()

	res, err := db.client.NewRaw(
		"INSERT INTO ? (id, user_id, order_id, status, created_at) VALUES (?, ?, ?, ?, ?)",
		bun.Ident("payments"), id, input.UserID, input.OrderID, input.Status, now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create Payment. Error: ", err)
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not retreive rows affected by Payment. ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) UpdatePayment(input model.PaymentInput) (bool, error) {

	row, err := db.client.NewRaw(
		"UPDATE payments SET user_id = ?, status = ? WHERE order_id = ?",
		input.UserID, input.Status, input.OrderID,
	).Exec(context.Background())
	if err != nil {
		fmt.Println("Could not update payment: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in update payment: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetPaymentBySession(session_id string) (*model.Payment, error) {
	var result model.Payment

	err := db.client.NewRaw("SELECT * FROM payments WHERE session_id = ?", session_id).Scan(context.Background(), &result)

	if err != nil {
		fmt.Print("Error found when querying for payments with id: ", err)
		return nil, err
	}

	return &result, nil
}
