package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"stockpilot/internal/domain"
	"stockpilot/internal/repository/dto"
	"stockpilot/pkg/gonerve/errors"
	"stockpilot/pkg/gonerve/genuuid"
	"stockpilot/pkg/gonerve/postgresql"
	"stockpilot/pkg/gonerve/postgresql/query"
)

type Repository struct {
	*postgresql.Repository
	ug genuuid.GeneratorUUID
}

func New(ctx context.Context, cfg postgresql.Config) (*Repository, error) {
	r, err := postgresql.NewRepository(ctx, cfg.ToConnString(), postgresql.WithListenNotifications(cfg.ListenNotifications))
	if err != nil {
		return nil, errors.Wrap(err, "postgresql.NewRepository")
	}
	return &Repository{Repository: r, ug: genuuid.New()}, nil
}

const createUserQuery = `
INSERT INTO users (id, email, first_name, last_name, age, is_married, password_hash, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, email, first_name, last_name, age, is_married, password_hash, created_at
`

func (r *Repository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if err := r.Locked(); err != nil {
		return nil, err
	}
	if user.ID == "" {
		user.ID = r.ug.V4()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	dbUser := dto.UserFromDomain(*user)
	conv := func(u dto.DBUser) (domain.User, error) {
		return dto.UserToDomain(u), nil
	}
	u, err := query.SelectOneWithConverterError(ctx, r.Conn, createUserQuery, conv, dbUser.ID, dbUser.Email, dbUser.FirstName, dbUser.LastName, dbUser.Age, dbUser.IsMarried, dbUser.PasswordHash, dbUser.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "create user")
	}
	return &u, nil
}

const getUserByEmailQuery = `
SELECT id, email, first_name, last_name, age, is_married, password_hash, created_at
FROM users
WHERE email = $1
`

func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	conv := func(u dto.DBUser) (domain.User, error) {
		return dto.UserToDomain(u), nil
	}
	u, err := query.SelectOneWithConverterError(ctx, r.Conn, getUserByEmailQuery, conv, email)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "get user by email")
	}
	return &u, nil
}

const getUserByIDQuery = `
SELECT id, email, first_name, last_name, age, is_married, password_hash, created_at
FROM users
WHERE id = $1
`

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	conv := func(u dto.DBUser) (domain.User, error) {
		return dto.UserToDomain(u), nil
	}
	u, err := query.SelectOneWithConverterError(ctx, r.Conn, getUserByIDQuery, conv, id)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "get user by id")
	}
	return &u, nil
}

const createProductQuery = `
INSERT INTO products (id, description, tags, quantity, price, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $6)
RETURNING id, description, tags, quantity, price, created_at, updated_at
`

func (r *Repository) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if err := r.Locked(); err != nil {
		return nil, err
	}
	if product.ID == "" {
		product.ID = r.ug.V4()
	}
	now := time.Now().UTC()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = now
	}
	dbProduct := dto.ProductFromDomain(*product)
	conv := func(p dto.DBProduct) (domain.Product, error) {
		return dto.ProductToDomain(p), nil
	}
	p, err := query.SelectOneWithConverterError(ctx, r.Conn, createProductQuery, conv, dbProduct.ID, dbProduct.Description, dbProduct.Tags, dbProduct.Quantity, dbProduct.Price, dbProduct.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "create product")
	}
	return &p, nil
}

const getProductByIDQuery = `
SELECT id, description, tags, quantity, price, created_at, updated_at
FROM products
WHERE id = $1
`

func (r *Repository) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	conv := func(p dto.DBProduct) (domain.Product, error) {
		return dto.ProductToDomain(p), nil
	}
	p, err := query.SelectOneWithConverterError(ctx, r.Conn, getProductByIDQuery, conv, id)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "get product by id")
	}
	return &p, nil
}

const getProductsForUpdateQuery = `
SELECT id, description, tags, quantity, price, created_at, updated_at
FROM products
WHERE id = ANY($1)
FOR UPDATE
`

func (r *Repository) GetByIDsForUpdate(ctx context.Context, tx pgx.Tx, ids []string) ([]domain.Product, error) {
	items, err := query.GetAll[dto.DBProduct](ctx, tx, getProductsForUpdateQuery, ids)
	if err != nil {
		return nil, errors.Wrap(err, "get products for update")
	}
	result := make([]domain.Product, 0, len(items))
	for _, p := range items {
		result = append(result, dto.ProductToDomain(p))
	}
	return result, nil
}

const updateQuantityQuery = `
UPDATE products
SET quantity = quantity + $2, updated_at = $3
WHERE id = $1 AND quantity + $2 >= 0
RETURNING id
`

func (r *Repository) UpdateQuantity(ctx context.Context, tx pgx.Tx, id string, delta int) error {
	err := query.Exec(ctx, tx, updateQuantityQuery, id, delta, time.Now().UTC())
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return errors.New("insufficient stock")
		}
		return errors.Wrap(err, "update quantity")
	}
	return nil
}

const createOrderQuery = `
INSERT INTO orders (id, user_id, created_at, total_price)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, created_at, total_price
`

func (r *Repository) CreateOrder(ctx context.Context, tx pgx.Tx, order *domain.Order, items []domain.OrderItem) (*domain.Order, error) {
	if order.ID == "" {
		order.ID = r.ug.V4()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	dbOrder := dto.OrderFromDomain(*order)
	var inserted dto.DBOrder
	err := tx.QueryRow(ctx, createOrderQuery, dbOrder.ID, dbOrder.UserID, dbOrder.CreatedAt, dbOrder.TotalPrice).Scan(&inserted.ID, &inserted.UserID, &inserted.CreatedAt, &inserted.TotalPrice)
	if err != nil {
		return nil, errors.Wrap(err, "insert order")
	}
	for i := range items {
		if items[i].ID == "" {
			items[i].ID = r.ug.V4()
		}
		items[i].OrderID = inserted.ID
		dbItem := dto.OrderItemFromDomain(items[i])
		if err := query.Exec(ctx, tx, `
INSERT INTO order_items (id, order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4, $5)
`, dbItem.ID, dbItem.OrderID, dbItem.ProductID, dbItem.Quantity, dbItem.Price); err != nil {
			return nil, errors.Wrap(err, "insert order item")
		}
	}
	conv := dto.OrderToDomain(inserted, items)
	return &conv, nil
}
