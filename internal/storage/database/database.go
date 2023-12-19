package database

import (
	"context"
	"database/sql"

	"github.com/The-Gleb/loyalty-system/internal/models"
)

var (
	//go:embed sqls/schema.sql
	schemaQuery string
)

type DB struct {
	db *sql.DB
}

func ConnectDB(dsn string) (*DB, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		// return nil, checkForConectionErr("ConnectDB", err)
	}

	// schemaQuery = strings.TrimSpace(schemaQuery)
	// logger.Log.Info(schemaQuery)
	_, err = db.ExecContext(context.Background(), schemaQuery)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) CreateUser(ctx context.Context, user models.Credentials) error {

}
func (db *DB) CreateSession(ctx context.Context, session models.Session) error {

}
func (db *DB) GetUserOrders(ctx context.Context, userName string) ([]models.Order, error) {

}
func (db *DB) GetBalance(ctx context.Context, userName string) (models.Balance, error) {

}
func (db *DB) TopUpBalance(ctx context.Context, orderNumber string, amountToAdd int) error {

}
func (db *DB) GetWithdrawalsInfo(ctx context.Context, userName string) ([]models.Withdrawal, error) {

}
func (db *DB) Withdraw(ctx context.Context, user string, withdrawal models.Withdrawal) error {

}
func (db *DB) AddWithdrawal(ctx context.Context, user string, withdrawal models.Withdrawal) error {

}
func (db *DB) AddOrder(ctx context.Context, user, orderNumber string) error {

}
func (db *DB) ChangeOrderStatus(ctx context.Context, orderNumber, newStatus string) error {

}
func (db *DB) GetNotProcessedOrders(ctx context.Context, user string) ([]models.Order, error) {

}
func (db *DB) GetUserPassword(ctx context.Context, login string) (string, error) {

}
