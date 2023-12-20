package database

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
	"time"

	"github.com/The-Gleb/loyalty-system/internal/models"
	_ "github.com/jackc/pgx/v5"
)

var (
	//go:embed sql/schema.sql
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

	schemaQuery = strings.TrimSpace(schemaQuery)
	// logger.Log.Info(schemaQuery)
	_, err = db.ExecContext(context.Background(), schemaQuery)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) CreateUser(ctx context.Context, user models.Credentials) error {

	_, err := db.db.ExecContext(ctx, `
				INSERT INTO users (login, password, current, withdrawn)
				VALUES ($1, $2, $3, $4);
			`, user.Login, user.Password, 0, 0)

	if err != nil {
		// TODO
	}

	return nil

}
func (db *DB) CreateSession(ctx context.Context, session models.Session) error {

	_, err := db.db.ExecContext(ctx, `
				INSERT INTO sessions (user, session_token, expiry)
				VALUES ($1, $2, $3);
			`, session.UserName, session.Token, session.Expiry)
	// TODO: handle timestamp
	if err != nil {
		// TODO
	}

	return nil

}
func (db *DB) GetUserOrders(ctx context.Context, userName string) ([]models.Order, error) {

	rows, err := db.db.QueryContext(ctx, `
				SELECT * FROM orders
				WHERE order_user = $1;
			`, userName)
	// TODO: handle timestamp
	defer rows.Close()
	if err != nil {
		// TODO
	}
	if rows.Err() != nil {
		// TODO
	}

	orders := make([]models.Order, 0)
	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.User, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			// TODO
		}
		orders = append(orders, order)
	}

	return orders, nil

}
func (db *DB) GetBalance(ctx context.Context, userName string) (models.Balance, error) {
	var balance models.Balance

	row := db.db.QueryRowContext(ctx, `
		SELECT current, withdrawn
		FROM users
		WHERE user = $1;
	`, userName)

	row.Scan(&balance.Current, &balance.Withdrawn)
	if row.Err() != nil {
		// TODO
	}

	return balance, nil

}
func (db *DB) TopUpBalance(ctx context.Context, orderNumber string, amountToAdd int) error {

	_, err := db.db.ExecContext(ctx, `
		UPDATE users
		SET current = current + $1
		WHERE login = $2
	`, amountToAdd, amountToAdd)
	if err != nil {
		// TODO
	}

	return nil

}
func (db *DB) GetWithdrawalsInfo(ctx context.Context, userName string) ([]models.Withdrawal, error) {

	withdrawals := make([]models.Withdrawal, 0)

	rows, err := db.db.QueryContext(ctx, `
		SEECT order, sum, processed_at
		FROM withdrawals
		WHERE user = $1
	`, userName)
	defer rows.Close()
	if err != nil {
		// TODO
	}

	for rows.Next() {
		var withdrawal models.Withdrawal
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			// TODO
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	if rows.Err() != nil {
		// TODO
	}

	return withdrawals, nil

}
func (db *DB) Withdraw(ctx context.Context, user string, withdrawal models.Withdrawal) error {

	tx, err := db.db.Begin()
	if err != nil {
		// TODO
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO withrawals (user, order, sum, processes_at)
		VALUES ($!, $2, $3, $4);
	`, user, withdrawal.Order, withdrawal.Sum, time.Now())
	// TODO: timestamp
	if err != nil {
		// TODO
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE users
		SET balance = balance - $1, withdrawn = withdrawn + $1
		WHERE user = $2;
	`, withdrawal.Sum, user)
	if err != nil {
		// TODO
	}
	tx.Commit()

	return nil

}

// func (db *DB) AddWithdrawal(ctx context.Context, user string, withdrawal models.Withdrawal) error {

// }
func (db *DB) AddOrder(ctx context.Context, user, orderNumber string) error {

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO withrawals (order_user, order_number, order_status, order_accrual, uploaded_at)
		VALUES ($1, $2, $3, $4, $5);
	`, user, orderNumber, "NEW", 0, time.Now())
	// TODO: timestamp
	if err != nil {
		// TODO
	}

	return nil

}
func (db *DB) UpdateOrder(ctx context.Context, order models.Order) error {

	_, err := db.db.ExecContext(ctx, `
	UPDATE orders
	SET order_status = $1,
		order_accrual = $2
	WHERE order_number = $3
`, order.Status, order.Accrual, order.Number)
	// TODO: timestamp
	if err != nil {
		// TODO
	}

	return nil

}
func (db *DB) GetNotProcessedOrders(ctx context.Context, user string) ([]models.Order, error) {

	rows, err := db.db.QueryContext(ctx, `
				SELECT * FROM orders
				WHERE order_user = $1"
				&& order_status IN ('NEW', 'PROCESSING');
			`, user)
	// TODO: handle timestamp
	defer rows.Close()
	if err != nil {
		// TODO
	}
	if rows.Err() != nil {
		// TODO
	}

	orders := make([]models.Order, 0)
	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.User, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			// TODO
		}
		orders = append(orders, order)
	}

	return orders, nil

}
func (db *DB) GetUserPassword(ctx context.Context, login string) (string, error) {

	row := db.db.QueryRowContext(ctx, `
		SELECT password FROM users
		WHERE login = $1
	`, login)
	if row.Err() != nil {
		// TODO
	}

	var pw string
	err := row.Scan(&pw)
	if err != nil {
		// TODO
	}

	return pw, nil

}
