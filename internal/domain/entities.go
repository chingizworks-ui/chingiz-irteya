package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID           string
	Email        string
	FirstName    string
	LastName     string
	Age          int
	IsMarried    bool
	PasswordHash string
	CreatedAt    time.Time
}

func (u User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return ""
	}
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

type Product struct {
	ID          string
	Description string
	Tags        []string
	Quantity    int
	Price       decimal.Decimal
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Order struct {
	ID         string
	UserID     string
	CreatedAt  time.Time
	TotalPrice decimal.Decimal
	Items      []OrderItem
}

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	Price     decimal.Decimal
}
