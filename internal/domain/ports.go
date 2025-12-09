package domain

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *Product) (*Product, error)
	GetProductByID(ctx context.Context, id string) (*Product, error)
	GetByIDsForUpdate(ctx context.Context, tx pgx.Tx, ids []string) ([]Product, error)
	UpdateQuantity(ctx context.Context, tx pgx.Tx, id string, delta int) error
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, tx pgx.Tx, order *Order, items []OrderItem) (*Order, error)
}

type TxManager interface {
	WithTx(ctx context.Context, f func(ctx context.Context, tx pgx.Tx) error) error
}
