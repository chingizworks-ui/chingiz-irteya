package tests

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"stockpilot/internal/domain"
	"stockpilot/pkg/gonerve/errors"
	"stockpilot/pkg/gonerve/genuuid"
)

type MemoryRepository struct {
	mu       sync.Mutex
	users    map[string]domain.User
	products map[string]domain.Product
	orders   map[string]domain.Order
	ug       genuuid.GeneratorUUID
}

type memoryTx struct {
	repo *MemoryRepository
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:    map[string]domain.User{},
		products: map[string]domain.Product{},
		orders:   map[string]domain.Order{},
		ug:       genuuid.New(),
	}
}

func (r *MemoryRepository) lock(tx pgx.Tx) func() {
	if mt, ok := tx.(*memoryTx); ok && mt.repo == r {
		return func() {}
	}
	r.mu.Lock()
	return r.mu.Unlock
}

func (r *MemoryRepository) nextID() string {
	return r.ug.V4()
}

func (r *MemoryRepository) WithTx(ctx context.Context, f func(ctx context.Context, tx pgx.Tx) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return f(ctx, &memoryTx{repo: r})
}

func (r *MemoryRepository) CreateUser(_ context.Context, user *domain.User) (*domain.User, error) {
	unlock := r.lock(nil)
	defer unlock()

	for _, u := range r.users {
		if strings.EqualFold(u.Email, user.Email) {
			return nil, errors.New("user already exists")
		}
	}

	if user.ID == "" {
		user.ID = r.nextID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	clone := *user
	r.users[user.ID] = clone
	return &clone, nil
}

func (r *MemoryRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	unlock := r.lock(nil)
	defer unlock()

	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			clone := u
			return &clone, nil
		}
	}
	return nil, nil
}

func (r *MemoryRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	unlock := r.lock(nil)
	defer unlock()

	if u, ok := r.users[id]; ok {
		clone := u
		return &clone, nil
	}
	return nil, nil
}

func (r *MemoryRepository) CreateProduct(_ context.Context, product *domain.Product) (*domain.Product, error) {
	unlock := r.lock(nil)
	defer unlock()

	if product.ID == "" {
		product.ID = r.nextID()
	}
	now := time.Now().UTC()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = now
	}
	clone := *product
	r.products[product.ID] = clone
	return &clone, nil
}

func (r *MemoryRepository) GetProductByID(_ context.Context, id string) (*domain.Product, error) {
	unlock := r.lock(nil)
	defer unlock()

	if p, ok := r.products[id]; ok {
		clone := p
		return &clone, nil
	}
	return nil, nil
}

func (r *MemoryRepository) GetByIDsForUpdate(_ context.Context, tx pgx.Tx, ids []string) ([]domain.Product, error) {
	unlock := r.lock(tx)
	defer unlock()

	result := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := r.products[id]; ok {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *MemoryRepository) UpdateQuantity(_ context.Context, tx pgx.Tx, id string, delta int) error {
	unlock := r.lock(tx)
	defer unlock()

	p, ok := r.products[id]
	if !ok {
		return errors.New("product not found")
	}
	if p.Quantity+delta < 0 {
		return errors.New("insufficient stock")
	}
	p.Quantity += delta
	p.UpdatedAt = time.Now().UTC()
	r.products[id] = p
	return nil
}

func (r *MemoryRepository) CreateOrder(_ context.Context, tx pgx.Tx, order *domain.Order, items []domain.OrderItem) (*domain.Order, error) {
	unlock := r.lock(tx)
	defer unlock()

	if order.ID == "" {
		order.ID = r.nextID()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	order.Items = make([]domain.OrderItem, len(items))
	copy(order.Items, items)

	for i := range order.Items {
		if order.Items[i].ID == "" {
			order.Items[i].ID = r.nextID()
		}
		order.Items[i].OrderID = order.ID
	}

	clone := *order
	r.orders[order.ID] = clone
	return &clone, nil
}

func (memoryTx) Begin(ctx context.Context) (pgx.Tx, error) { return memoryTx{}, nil }
func (memoryTx) Commit(ctx context.Context) error          { return nil }
func (memoryTx) Rollback(ctx context.Context) error        { return nil }
func (memoryTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (memoryTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	var br pgx.BatchResults
	return br
}
func (memoryTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (memoryTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return &pgconn.StatementDescription{}, nil
}
func (memoryTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (memoryTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (memoryTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (memoryTx) Conn() *pgx.Conn                                               { return nil }
