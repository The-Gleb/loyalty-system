package models

import "time"

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Session struct {
	UserName string
	Token    string
	Expiry   time.Time
}

func (s *Session) isExpired() bool {
	return s.Expiry.Before(time.Now())
}

type Order struct {
	User       string    `json:"user,omitempty"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"prosecced_at,omitempty"`
}
