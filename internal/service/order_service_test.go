package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"stockpilot/internal/domain"
	"stockpilot/pkg/gonerve/errors"
)

type txMock struct{}

func (txMock) Begin(ctx context.Context) (pgx.Tx, error) { return txMock{}, nil }
func (txMock) Commit(ctx context.Context) error          { return nil }
func (txMock) Rollback(ctx context.Context) error        { return nil }
func (txMock) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (txMock) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	var br pgx.BatchResults
	return br
}
func (txMock) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (txMock) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return &pgconn.StatementDescription{}, nil
}
func (txMock) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (txMock) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) { return nil, nil }
func (txMock) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row        { return nil }
func (txMock) Conn() *pgx.Conn                                                      { return nil }

type txManagerMock struct {
	tx pgx.Tx
}

func (m txManagerMock) WithTx(ctx context.Context, f func(ctx context.Context, tx pgx.Tx) error) error {
	return f(ctx, m.tx)
}

type productRepoMock struct {
	items map[string]domain.Product
}

func (m *productRepoMock) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	return product, nil
}

func (m *productRepoMock) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	if p, ok := m.items[id]; ok {
		return &p, nil
	}
	return nil, nil
}

func (m *productRepoMock) GetByIDsForUpdate(ctx context.Context, tx pgx.Tx, ids []string) ([]domain.Product, error) {
	result := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := m.items[id]; ok {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *productRepoMock) UpdateQuantity(ctx context.Context, tx pgx.Tx, id string, delta int) error {
	p, ok := m.items[id]
	if !ok {
		return errors.New("product not found")
	}
	if p.Quantity+delta < 0 {
		return errors.New("insufficient stock")
	}
	p.Quantity += delta
	m.items[id] = p
	return nil
}

type orderRepoMock struct {
	created *domain.Order
}

func (m *orderRepoMock) CreateOrder(ctx context.Context, tx pgx.Tx, order *domain.Order, items []domain.OrderItem) (*domain.Order, error) {
	o := *order
	o.Items = items
	m.created = &o
	return m.created, nil
}

type orderUserRepoMock struct {
	user *domain.User
}

func (orderUserRepoMock) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return nil, nil
}

func (m orderUserRepoMock) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (m orderUserRepoMock) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.user, nil
}

func TestOrderCreateInsufficientStock(t *testing.T) {
	products := &productRepoMock{
		items: map[string]domain.Product{
			"p1": {ID: "p1", Quantity: 1, Price: decimal.NewFromInt(10)},
		},
	}
	orders := &orderRepoMock{}
	users := orderUserRepoMock{user: &domain.User{ID: "u1"}}
	svc := NewOrderService(products, orders, users, txManagerMock{tx: txMock{}})

	_, err := svc.Create(context.Background(), "u1", []OrderItemInput{
		{ProductID: "p1", Quantity: 2},
	})
	require.Error(t, err)
	require.Nil(t, orders.created)
}

func TestOrderCreateSuccess(t *testing.T) {
	products := &productRepoMock{
		items: map[string]domain.Product{
			"p1": {ID: "p1", Quantity: 5, Price: decimal.NewFromInt(15)},
		},
	}
	orders := &orderRepoMock{}
	users := orderUserRepoMock{user: &domain.User{ID: "u1"}}
	svc := NewOrderService(products, orders, users, txManagerMock{tx: txMock{}})

	order, err := svc.Create(context.Background(), "u1", []OrderItemInput{
		{ProductID: "p1", Quantity: 2},
	})
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, decimal.NewFromInt(30), order.TotalPrice)
	require.Len(t, order.Items, 1)
	require.Equal(t, 3, products.items["p1"].Quantity)
	require.Equal(t, "p1", order.Items[0].ProductID)
	require.Equal(t, 2, order.Items[0].Quantity)
	require.True(t, order.Items[0].Price.Equal(decimal.NewFromInt(15)))
}
