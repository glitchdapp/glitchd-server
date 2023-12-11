package database

import (
	"context"
	"fmt"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
)

func (db *BUN) CreatePage(input model.NewPage) (*model.Page, error) {
	id := uuid.New()
	data := model.Page{
		ID:          id.String(),
		AppID:       input.AppID,
		Title:       input.Title,
		Description: input.Description,
		Type:        input.Type,
	}

	res, err := db.client.NewInsert().Model(&data).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create page. Error: ", err)
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Error fetching rows after inserting page into db: ", err)
		return &model.Page{}, err
	}

	if row > 0 {
		fmt.Println("Created page successfully: ", input.Title)
		return &data, nil
	}

	return &data, nil
}

func (db *BUN) UpdatePage(id string, input model.UpdatePage) (bool, error) {
	var page model.Page

	res, err := db.client.NewUpdate().Model(&page).Set("title = ?", input.Title).Where("id = ?", id).Exec(context.Background())

	if err != nil {
		fmt.Println("Error updating page: ", err)
		return false, err
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Error fetching rows after updating page: ", err)
		return false, err
	}

	if row > 0 {
		fmt.Println("Updated page successfully: ", input.Title)
		return true, nil
	}

	return false, nil
}

func (db *BUN) DeletePage(id string) (bool, error) {
	var page model.Page

	res, err := db.client.NewDelete().Model(&page).Where("id = ?", id).Exec(context.Background())

	if err != nil {
		fmt.Println("Error deleting page: ", err)
		return false, err
	}

	row, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Error fetching rows after deleting page: ", err)
		return false, err
	}

	if row > 0 {
		fmt.Println("Deleted page successfully: ", id)
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetPages(app_id string) ([]*model.Page, error) {
	var page []*model.Page

	err := db.client.NewSelect().Model(&page).Where("app_id = ?", app_id).Scan(context.Background())

	if err != nil {
		fmt.Println("Error fetching page: ", err)
		return nil, err
	}

	return page, nil
}

func (db *BUN) GetPage(id string) (*model.Page, error) {
	var page *model.Page

	err := db.client.NewSelect().Model(&page).Where("id = ?", id).Scan(context.Background())

	if err != nil {
		fmt.Println("Error fetching page: ", err)
		return nil, err
	}

	return page, nil
}
