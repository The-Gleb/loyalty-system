package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/The-Gleb/loyalty-system/internal/errors"
)

type App interface {
	Register(ctx context.Context, body io.ReadCloser) (string, time.Time, error)
	Login(ctx context.Context, body io.ReadCloser) (string, time.Time, error)
	GetOrdersInfo(ctx context.Context, userName string) ([]byte, error)
	LoadOrder(ctx context.Context, user string, orderNumber io.ReadCloser) error
	GetBalance(ctx context.Context, user string) ([]byte, error)
	Withdraw(ctx context.Context, user string, body io.ReadCloser) error
	GetWithdrawalsInfo(ctx context.Context, user string) ([]byte, error)
}

type handlers struct {
	app App
}

func New(app App) *handlers {
	return &handlers{
		app: app,
	}
}

func (h *handlers) RegisterHandler(rw http.ResponseWriter, r *http.Request) {

	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(rw, err.Error(), http.StatusInternalServerError)
	// }
	sessionToken, expires, err := h.app.Register(r.Context(), r.Body)

	if err != nil {
		switch errors.Code(err) {
		case errors.LoginAlredyExists:
			http.Error(rw, err.Error(), http.StatusConflict)
			return
		case errors.ErrUnmarshallingJSON:
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	c := http.Cookie{
		Name:    "SESSION_TOKEN",
		Value:   sessionToken,
		Expires: expires,
	}

	http.SetCookie(rw, &c)

}
func (h *handlers) LoginHandler(rw http.ResponseWriter, r *http.Request) {

	sessionToken, expires, err := h.app.Login(r.Context(), r.Body)

	if err != nil {
		switch errors.Code(err) {
		case errors.WrongLoginOrPassword:
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		case errors.ErrUnmarshallingJSON:
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	c := http.Cookie{
		Name:    "SESSION_TOKEN",
		Value:   sessionToken,
		Expires: expires,
	}

	http.SetCookie(rw, &c)

}
func (h *handlers) GetOrdersInfoHandler(rw http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value("user").(string)
	if !ok {
		http.Error(rw, "couldn`t get user name from req context", http.StatusInternalServerError)
	}
	body, err := h.app.GetOrdersInfo(r.Context(), user)

	if err != nil {
		switch errors.Code(err) {
		case errors.NoDataFound:
			rw.WriteHeader(204)
			return
		default:
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	rw.Write(body)

}
func (h *handlers) LoadOrderHandler(rw http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value("user").(string)
	if !ok {
		http.Error(rw, "couldn`t get user name from req context", http.StatusInternalServerError)
	}

	err := h.app.LoadOrder(r.Context(), user, r.Body)

	if err != nil {
		switch errors.Code(err) {
		case errors.OrderAlreadyAddedByThisUser:
			rw.WriteHeader(http.StatusOK)
			return
		case errors.ErrUnmarshallingJSON:
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		case errors.OrderAlreadyAddedByAnotherUser:
			http.Error(rw, err.Error(), http.StatusConflict)
			return
		case errors.InvalidOrderNumber:
			http.Error(rw, err.Error(), http.StatusUnprocessableEntity)
			return
		default:
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	rw.WriteHeader(202)

}
func (h *handlers) GetBalanceHandler(rw http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value("user").(string)
	if !ok {
		http.Error(rw, "couldn`t get user name from req context", http.StatusInternalServerError)
	}

	body, err := h.app.GetBalance(r.Context(), user)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	rw.Write(body)

}
func (h *handlers) WithdrawHandler(rw http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value("user").(string)
	if !ok {
		http.Error(rw, "couldn`t get user name from req context", http.StatusInternalServerError)
	}

	err := h.app.Withdraw(r.Context(), user, r.Body)
	if err != nil {
		switch errors.Code(err) {
		case errors.InsufficientFunds:
			http.Error(rw, err.Error(), http.StatusPaymentRequired)
			return
		case errors.InvalidOrderNumber:
			http.Error(rw, err.Error(), http.StatusUnprocessableEntity)
			return
		default:
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}
func (h *handlers) GetWithdrawalsInfoHandler(rw http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value("user").(string)
	if !ok || user == "" {
		http.Error(rw, "couldn`t get user name from req context", http.StatusInternalServerError)
	}

	respBody, err := h.app.GetWithdrawalsInfo(r.Context(), user)
	if err != nil {
		switch errors.Code(err) {
		case errors.NoDataFound:
			rw.WriteHeader(204)
		default:
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	rw.Write(respBody)

}
