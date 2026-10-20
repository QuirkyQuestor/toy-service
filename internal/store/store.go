package store

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/QuirkyQuestor/toy-service/internal/config"
	"github.com/QuirkyQuestor/toy-service/internal/datamodels"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrSQLExecution                = errors.New("error during sql statement execution")
	ErrCannotPrepareSQLStatement   = errors.New("cannot prepare sql statement")
	ErrConvertingDBResponse        = errors.New("error during converting DB/Go types")
	ErrNoItemResponse              = errors.New("DB query returned no result")
	ErrUnexpectedDBExecutionResult = errors.New("unexpected DB statement ExecutionResult")
)

func DBConnect(stopFunc context.CancelFunc) (*sql.DB, error) {
	dbSettings := config.GetDBConfig(stopFunc)
	if dbSettings == nil {
		return nil, errors.New("could not find DB config")
	}
	db, err := sql.Open("pgx", dbSettings.DATABASE_URL)
	if err != nil {
		slog.Error("could not open DB Connection", "error", err)
		return nil, err
	}

	return db, nil
}

func GetAllItems(ctx context.Context, db *sql.DB) ([]datamodels.Item, error) {
	allItems := []datamodels.Item{}

	query := `SELECT id, name FROM items;`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		slog.Error(ErrSQLExecution.Error(), "error", err)
		return nil, ErrSQLExecution
	}
	defer rows.Close()
	for rows.Next() {
		item := datamodels.Item{}
		err := rows.Scan(&item.Id, &item.Name)
		if err != nil {
			slog.Error("Could not parse row from the DB", "error", err)
			return allItems, err
		}
		allItems = append(allItems, item)
	}

	if err := rows.Err(); err != nil {
		return allItems, err
	}

	return allItems, nil
}

func GetItemById(ctx context.Context, db *sql.DB, id int64) (*datamodels.Item, error) {

	query := `SELECT id, name FROM items WHERE id = $1;`

	item := datamodels.Item{}
	err := db.QueryRowContext(ctx, query, id).Scan(&item.Id, &item.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info(ErrNoItemResponse.Error())
			return nil, ErrNoItemResponse
		}
		slog.Error(ErrConvertingDBResponse.Error(), "error", err)
		return nil, ErrConvertingDBResponse
	}

	return &item, nil
}

func AddNewItem(ctx context.Context, db *sql.DB, name string) (*datamodels.Item, error) {
	query := `INSERT INTO items(name) VALUES ($1) RETURNING id;`

	item := &datamodels.Item{
		Name: name,
	}

	err := db.QueryRowContext(ctx, query, name).Scan(&item.Id)

	if err != nil {
		slog.Error(ErrSQLExecution.Error(), "error", err)
		return item, ErrSQLExecution

	}

	return item, nil
}

func DeleteItemById(ctx context.Context, db *sql.DB, id int64) error {
	query := `DELETE FROM items WHERE id = $1;`

	result, err := db.ExecContext(ctx, query, id)

	if err != nil {
		slog.Error(ErrSQLExecution.Error(), "error", err)
		return ErrSQLExecution
	}

	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		slog.Error(ErrUnexpectedDBExecutionResult.Error(), "error", err)
		return ErrUnexpectedDBExecutionResult
	}
	if rowsUpdated == 0 {
		return ErrNoItemResponse
	}

	return nil
}
