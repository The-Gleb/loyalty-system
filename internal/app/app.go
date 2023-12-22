package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/The-Gleb/loyalty-system/internal/errors"
	"github.com/The-Gleb/loyalty-system/internal/models"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUserPassword(ctx context.Context, login string) (string, error)
	CreateUser(ctx context.Context, user models.Credentials) error
	CreateSession(ctx context.Context, session models.Session) error
	GetUserOrders(ctx context.Context, userName string) ([]models.Order, error)
	GetBalance(ctx context.Context, userName string) (models.Balance, error)
	TopUpBalance(ctx context.Context, orderNumber string, amountToAdd int) error
	GetWithdrawalsInfo(ctx context.Context, userName string) ([]models.Withdrawal, error)
	Withdraw(ctx context.Context, user string, withdrawal models.Withdrawal) error
	// AddWithdrawal(ctx context.Context, user string, withdrawal models.Withdrawal) error
	// GetOrdersInfo(ctx context.Context, orderNumber string) error
	AddOrder(ctx context.Context, user, orderNumber string) (models.Order, error)
	UpdateOrder(ctx context.Context, order models.Order) error
	GetNotProcessedOrders(ctx context.Context, user string) ([]models.Order, error)
}

type app struct {
	client  *resty.Client
	storage Repository
}

func NewApp(s Repository, accrualAddr string) *app {
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryMaxWaitTime(60 * time.Second).
		AddRetryCondition(func(c *resty.Response, err error) bool {
			return c.StatusCode() == 429
		}).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			var dur time.Duration
			if r.StatusCode() == 429 {
				c.SetRetryCount(c.RetryCount + 1)
				dur = time.Duration(60) * time.Second
			} else {
				dur = time.Duration(r.Request.Attempt*2-1) * time.Second
			}
			// log.Printf("attempt: %d", r.Request.Attempt)

			return dur, nil
		}).
		SetBaseURL(accrualAddr)
	return &app{
		client:  client,
		storage: s,
	}
}

func (a *app) CheckOrderAccrual(ctx context.Context, orderNumber []byte) (models.Order, error) {

	resp, err := a.client.R().
		Post("/orders/" + string(orderNumber))
	if err != nil {
		// TODO
	}

	var order models.Order
	err = json.Unmarshal(resp.Body(), &order)
	if err != nil {
		// TODO
	}

	if order.Status == "REGISTERED" {
		order.Status = "NEW"
		return order, nil
	}

	err = a.storage.UpdateOrder(ctx, order)
	if err != nil {
		// TODO
	}

	err = a.storage.TopUpBalance(ctx, string(orderNumber), order.Accrual)
	if err != nil {
		// TODO
	}

	return order, nil

}

func isValid(orderNumber []byte) bool {
	// TODO
	return true
}

func (a *app) Register(ctx context.Context, body io.ReadCloser) (string, time.Time, error) {
	var newUser models.Credentials

	err := json.NewDecoder(body).Decode(&newUser)
	if err != nil {
		return "", time.Now(), errors.WrapIntoDomainError(err, errors.ErrUnmarshallingJSON, "[Register]:")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", time.Now(), fmt.Errorf("[Register]: %w", err)
	}
	newUser.Password = string(hashedPassword)

	err = a.storage.CreateUser(ctx, newUser)
	if err != nil {
		return "", time.Now(), fmt.Errorf("[Register]: %w", err)
	}

	newSession := models.Session{
		UserName: newUser.Login,
		Token:    uuid.NewString(),
		Expiry:   time.Now().Add(24 * time.Hour),
	}

	for {
		err = a.storage.CreateSession(ctx, newSession)
		if errors.Code(err) == errors.NotUniqueToken {
			newSession.Token = uuid.NewString()
			continue
		}
		if err != nil {
			return "", time.Now(), fmt.Errorf("[Register]: %w", err)
		}
		break
	}

	return newSession.Token, newSession.Expiry, nil

}
func (a *app) Login(ctx context.Context, body io.ReadCloser) (string, time.Time, error) {

	var user models.Credentials

	err := json.NewDecoder(body).Decode(&user)
	if err != nil {
		return "", time.Now(), errors.WrapIntoDomainError(err, errors.ErrUnmarshallingJSON, "[Register]:")
	}

	expectedPassword, err := a.storage.GetUserPassword(ctx, user.Login)
	if err != nil {
		return "", time.Now(), fmt.Errorf("[Login]: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(expectedPassword), []byte(user.Password))
	if err != nil {
		return "", time.Now(), errors.WrapIntoDomainError(err, errors.WrongLoginOrPassword, "[Login]:")
	}

	newSession := models.Session{
		UserName: user.Login,
		Token:    uuid.NewString(),
		Expiry:   time.Now().Add(24 * time.Hour),
	}

	for {
		err = a.storage.CreateSession(ctx, newSession)
		if errors.Code(err) == errors.NotUniqueToken {
			newSession.Token = uuid.NewString()
			continue
		}
		if err != nil {
			return "", time.Now(), fmt.Errorf("[Register]: %w", err)
		}
		break
	}

	return newSession.Token, newSession.Expiry, nil

}
func (a *app) GetOrdersInfo(ctx context.Context, userName string) ([]byte, error) {

	notProcessedOrdersrders, err := a.storage.GetNotProcessedOrders(ctx, userName)

	var wg sync.WaitGroup
	for _, order := range notProcessedOrdersrders {
		wg.Add(1)
		go func(orderNumber []byte) {
			a.CheckOrderAccrual(ctx, orderNumber)
			wg.Done()
		}([]byte(order.Number))
	}

	wg.Wait()

	orders, err := a.storage.GetUserOrders(ctx, userName)
	if err != nil {
		// TODO
	}

	ordersJSON, err := json.Marshal(orders)
	if err != nil {
		// TODO
	}

	return ordersJSON, nil

}
func (a *app) LoadOrder(ctx context.Context, user string, orderNumber io.ReadCloser) error {

	orderNum, err := io.ReadAll(orderNumber)
	if err != nil {
		return err
	}

	if !isValid(orderNum) {
		return errors.NewDomainError(errors.InvalidOrderNumber, "[LoadOrder]: order number is invalid")
	}

	_, err = a.storage.AddOrder(ctx, user, string(orderNum))
	if err != nil {
		return err
	}

	go a.CheckOrderAccrual(ctx, orderNum)

	return nil

}
func (a *app) GetBalance(ctx context.Context, user string) ([]byte, error) {

	orders, err := a.storage.GetNotProcessedOrders(ctx, user)

	var wg sync.WaitGroup
	for _, order := range orders {
		wg.Add(1)
		go func(orderNumber []byte) {
			a.CheckOrderAccrual(ctx, orderNumber)
			wg.Done()
		}([]byte(order.Number))
	}

	wg.Wait()

	balance, err := a.storage.GetBalance(ctx, user)
	if err != nil {
		return make([]byte, 0), err
	}

	jsonBalance, err := json.Marshal(balance)
	if err != nil {
		return make([]byte, 0), err
	}

	return jsonBalance, nil

}
func (a *app) Withdraw(ctx context.Context, user string, body io.ReadCloser) error {

	var withdrawRequest models.Withdrawal

	json.NewDecoder(body).Decode(&withdrawRequest)

	balance, err := a.storage.GetBalance(ctx, user)
	if err != nil {
		return err
	}

	if withdrawRequest.Sum > balance.Current {
		return errors.NewDomainError(errors.InsufficientFunds, "[Withdraw]: insufficient funds")
	}

	err = a.storage.Withdraw(ctx, user, withdrawRequest)
	if err != nil {
		return err
	}

	// withdrawRequest.ProcessedAt = time.Now()

	// err = a.storage.AddWithdrawal(ctx, user, withdrawRequest)

	return nil

}
func (a *app) GetWithdrawalsInfo(ctx context.Context, user string) ([]byte, error) {

	withdrawalsInfo, err := a.storage.GetWithdrawalsInfo(ctx, user)
	if err != nil {
		return make([]byte, 0), err
	}

	withdrawalsInfoJSON, err := json.Marshal(withdrawalsInfo)
	if err != nil {
		return make([]byte, 0), err
	}

	return withdrawalsInfoJSON, nil

}
