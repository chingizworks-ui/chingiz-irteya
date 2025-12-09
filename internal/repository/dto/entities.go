package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"stockpilot/internal/domain"
)

type DBUser struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	Age          int       `db:"age"`
	IsMarried    bool      `db:"is_married"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type DBProduct struct {
	ID          string          `db:"id"`
	Description string          `db:"description"`
	Tags        []string        `db:"tags"`
	Quantity    int             `db:"quantity"`
	Price       decimal.Decimal `db:"price"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

type DBOrder struct {
	ID         string          `db:"id"`
	UserID     string          `db:"user_id"`
	CreatedAt  time.Time       `db:"created_at"`
	TotalPrice decimal.Decimal `db:"total_price"`
}

type DBOrderItem struct {
	ID        string          `db:"id"`
	OrderID   string          `db:"order_id"`
	ProductID string          `db:"product_id"`
	Quantity  int             `db:"quantity"`
	Price     decimal.Decimal `db:"price"`
}

func UserFromDomain(u domain.User) DBUser {
	return DBUser{
		ID:           u.ID,
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Age:          u.Age,
		IsMarried:    u.IsMarried,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

func UserToDomain(u DBUser) domain.User {
	return domain.User{
		ID:           u.ID,
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Age:          u.Age,
		IsMarried:    u.IsMarried,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

func ProductFromDomain(p domain.Product) DBProduct {
	return DBProduct{
		ID:          p.ID,
		Description: p.Description,
		Tags:        p.Tags,
		Quantity:    p.Quantity,
		Price:       p.Price,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ProductToDomain(p DBProduct) domain.Product {
	return domain.Product{
		ID:          p.ID,
		Description: p.Description,
		Tags:        p.Tags,
		Quantity:    p.Quantity,
		Price:       p.Price,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func OrderFromDomain(o domain.Order) DBOrder {
	return DBOrder{
		ID:         o.ID,
		UserID:     o.UserID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice,
	}
}

func OrderToDomain(o DBOrder, items []domain.OrderItem) domain.Order {
	return domain.Order{
		ID:         o.ID,
		UserID:     o.UserID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice,
		Items:      items,
	}
}

func OrderItemFromDomain(i domain.OrderItem) DBOrderItem {
	return DBOrderItem{
		ID:        i.ID,
		OrderID:   i.OrderID,
		ProductID: i.ProductID,
		Quantity:  i.Quantity,
		Price:     i.Price,
	}
}

func OrderItemToDomain(i DBOrderItem) domain.OrderItem {
	return domain.OrderItem{
		ID:        i.ID,
		OrderID:   i.OrderID,
		ProductID: i.ProductID,
		Quantity:  i.Quantity,
		Price:     i.Price,
	}
}
