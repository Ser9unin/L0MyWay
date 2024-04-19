package storage

import (
	"context"
	"database/sql"
	"encoding/json"

	model "github.com/Ser9unin/L0MyWay/pkg/model"
)

type DBOrder struct {
	OrderUID string          `json:"name"`
	Data     json.RawMessage `json:"description"`
}

func UnmarshallNatsMsg(msg []byte) (DBOrder, error) {
	var dataModel model.Model

	err := json.Unmarshal(msg, &dataModel)
	if err != nil {
		return DBOrder{}, err
	}

	NewOrder := DBOrder{
		OrderUID: dataModel.OrderUID,
		Data:     msg,
	}

	return NewOrder, nil
}

const insertOrder = `-- name: InsertOrder :one
INSERT INTO models(order_uid, model) 
VALUES ($1, $2)
`

func InsertToDB(ctx context.Context, db *sql.DB, order DBOrder) error {
	_, err := db.ExecContext(ctx, insertOrder, order.OrderUID, order.Data)
	if err != nil {
		return err
	}
	return nil
}

const getOrderByID = `-- name: GetOrderByID :one
SELECT order_uid, model
FROM models
WHERE order_uid = $1
`

type GetOrderByIDRow struct {
	OrderUID string          `json:"order_uid"`
	Data     json.RawMessage `json:"data"`
}

func GetOrderByID(ctx context.Context, db *sql.DB, order DBOrder) (GetOrderByIDRow, error) {

	row := db.QueryRowContext(ctx, getOrderByID, order.OrderUID)
	var i GetOrderByIDRow
	err := row.Scan(&i.OrderUID, &i.Data)
	return i, err
}

const listOrders = `-- name: ListOrders :many
SELECT order_uid, model
FROM models
WHERE id > $1
LIMIT $2
`

type ListOrdersParams struct {
	ID    int64 `json:"id"`
	Limit int32 `json:"limit"`
}

type ListOrdersRow struct {
	OrderUID string          `json:"order_uid"`
	Data     json.RawMessage `json:"data"`
}

func ListOrders(ctx context.Context, db *sql.DB, arg ListOrdersParams) ([]ListOrdersRow, error) {
	rows, err := db.QueryContext(ctx, listOrders, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListOrdersRow
	for rows.Next() {
		var i ListOrdersRow
		if err := rows.Scan(&i.OrderUID, &i.Data); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
