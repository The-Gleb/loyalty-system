package database

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
	"time"

	"github.com/The-Gleb/loyalty-system/internal/errors"
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
		return nil, err
	}

	schemaQuery = strings.TrimSpace(schemaQuery)
	// logger.Log.Info(schemaQuery)
	_, err = db.ExecContext(context.Background(), schemaQuery)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// func handleErr(err error) error {
// 	if err
// }

func (db *DB) CreateUser(ctx context.Context, user models.Credentials) error {

	row := db.db.QueryRowContext(ctx, `
		SELECT login FROM users
		WHERE login = $1;
	`, user.Login)

	var login string
	err := row.Scan(&login)
	if err == nil {
		return errors.NewDomainError(errors.LoginAlredyExists, "[CreateUser]: login exists")
	}
	if err.Error() != sql.ErrNoRows.Error() {
		return err
	}

	_, err = db.db.ExecContext(ctx, `
				INSERT INTO users (login, password, current, withdrawn)
				VALUES ($1, $2, $3, $4);
			`, user.Login, user.Password, 0, 0)

	if err != nil {
		return err
	}

	return nil

}
func (db *DB) CreateSession(ctx context.Context, session models.Session) error {

	_, err := db.db.ExecContext(ctx, `
				INSERT INTO sessions (user, session_token, expiry)
				VALUES ($1, $2, $3);
			`, session.UserName, session.Token, session.Expiry)
	// TODO: handle timestamp
	// TODO: handle not unique token
	if err != nil {
		return err
	}

	return nil

}
func (db *DB) GetUserOrders(ctx context.Context, userName string) ([]models.Order, error) {

	rows, err := db.db.QueryContext(ctx, `
				SELECT * FROM orders
				WHERE order_user = $1;
			`, userName)
	// TODO: handle timestamp

	if err != nil {
		// if err == sql.ErrNoRows {
		// 	return make([]models.Order, 0), errors.WrapIntoDomainError(err, errors.NoDataFound, "[GetUserOrders]:")
		// }
		return make([]models.Order, 0), err
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.User, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				return make([]models.Order, 0), errors.WrapIntoDomainError(err, errors.NoDataFound, "[GetUserOrders]:")
			}
			return make([]models.Order, 0), err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		// if err == sql.ErrNoRows {
		// 	return make([]models.Order, 0), errors.WrapIntoDomainError(err, errors.NoDataFound, "[GetUserOrders]:")
		// }
		return make([]models.Order, 0), err
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
	if err := row.Err(); err != nil {
		return balance, err
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
		return err
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

	if err != nil {
		return make([]models.Withdrawal, 0), err
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal models.Withdrawal
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				return make([]models.Withdrawal, 0), errors.WrapIntoDomainError(err, errors.NoDataFound, "[GetUserOrders]:")
			}
			return make([]models.Withdrawal, 0), err
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	if rows.Err() != nil {
		return make([]models.Withdrawal, 0), err
	}

	return withdrawals, nil

}
func (db *DB) Withdraw(ctx context.Context, user string, withdrawal models.Withdrawal) error {

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO withrawals (user, order, sum, processes_at)
		VALUES ($1, $2, $3, $4);
	`, user, withdrawal.Order, withdrawal.Sum, time.Now())
	// TODO: timestamp
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE users
		SET balance = balance - $1, withdrawn = withdrawn + $1
		WHERE user = $2;
	`, withdrawal.Sum, user)
	if err != nil {
		return err
	}
	tx.Commit()

	return nil

}

// func (db *DB) AddWithdrawal(ctx context.Context, user string, withdrawal models.Withdrawal) error {

// }
func (db *DB) AddOrder(ctx context.Context, user, orderNumber string) (models.Order, error) {

	row := db.db.QueryRowContext(ctx, `
		SELECT * FROM orders
		WHERE order_number = $1;
	`, orderNumber)

	var order models.Order
	err := row.Scan(&order)
	if err == nil {
		if order.User == user {
			return order, errors.NewDomainError(errors.OrderAlreadyAddedByThisUser, "[AddOrder]:")
		}
		return order, errors.NewDomainError(errors.OrderAlreadyAddedByAnotherUser, "[AddOrder]:")
	}
	if err != sql.ErrNoRows {
		return order, err
	}

	_, err = db.db.ExecContext(ctx, `
		INSERT INTO withrawals (order_user, order_number, order_status, order_accrual, uploaded_at)
		VALUES ($1, $2, $3, $4, $5);
	`, user, orderNumber, "NEW", 0, time.Now())
	// TODO: timestamp
	if err != nil {
		return order, err
	}

	return models.Order{}, nil

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
		return err
	}

	return nil

}
func (db *DB) GetNotProcessedOrders(ctx context.Context, user string) ([]models.Order, error) {

	orders := make([]models.Order, 0)

	rows, err := db.db.QueryContext(ctx, `
				SELECT * FROM orders
				WHERE order_user = $1"
				&& order_status IN ('NEW', 'PROCESSING');
			`, user)
	// TODO: handle timestamp
	if err := rows.Err(); err != nil {
		return orders, err
	}
	if err != nil {
		return orders, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.User, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err == sql.ErrNoRows {
			return orders, nil
		}
		if err != nil {
			return orders, err
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
	if err := row.Err(); err != nil {
		return "", err
	}

	var pw string
	err := row.Scan(&pw)
	if err != nil {
		return "", err
	}

	return pw, nil

}
