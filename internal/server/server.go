package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

type Handlers interface {
	RegisterHandler(http.ResponseWriter, *http.Request)
	LoginHandler(http.ResponseWriter, *http.Request)
	GetOrdersInfoHandler(http.ResponseWriter, *http.Request)
	LoadOrderHandler(http.ResponseWriter, *http.Request)
	GetBalanceHandler(http.ResponseWriter, *http.Request)
	WithdrawHandler(http.ResponseWriter, *http.Request)
	GetWithdrawalsInfoHandler(http.ResponseWriter, *http.Request)
}

func New(address string, handlers Handlers) *http.Server {
	r := chi.NewRouter()
	setupRoutes(r, handlers)
	return &http.Server{
		Addr:    address,
		Handler: r,
	}
}

func setupRoutes(r *chi.Mux, h Handlers) {
	r.Post("/api/user/register", h.RegisterHandler)
	r.Post("/api/user/login", h.LoginHandler)
	r.Post("/api/user/orders", h.LoadOrderHandler)
	r.Get("/api/user/orders", h.GetOrdersInfoHandler)
	r.Get("/api/user/balance", h.GetBalanceHandler)
	r.Post("/api/user/balance/withdraw", h.WithdrawHandler)
	r.Get("/api/user/withdrawals", h.GetWithdrawalsInfoHandler)
}

func Run(s *http.Server) error {
	// logger.Log.Infow("Running server",
	// 	"address", s.Addr,
	// )
	return s.ListenAndServe()
}

func Shutdown(s *http.Server) {
	s.Shutdown(context.Background())
}
