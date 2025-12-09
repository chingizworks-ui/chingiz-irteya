package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"stockpilot/internal/domain"
	"stockpilot/pkg/gonerve/errors"
)

type OrderItemInput struct {
	ProductID string
	Quantity  int
}

type OrderService struct {
	products domain.ProductRepository
	orders   domain.OrderRepository
	users    domain.UserRepository
	tx       domain.TxManager
}

func NewOrderService(products domain.ProductRepository, orders domain.OrderRepository, users domain.UserRepository, tx domain.TxManager) *OrderService {
	return &OrderService{
		products: products,
		orders:   orders,
		users:    users,
		tx:       tx,
	}
}

func (s *OrderService) Create(ctx context.Context, userID string, items []OrderItemInput) (*domain.Order, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}
	if len(items) == 0 {
		return nil, errors.New("order items are required")
	}
	for _, item := range items {
		if item.ProductID == "" {
			return nil, errors.New("product id is required")
		}
		if item.Quantity <= 0 {
			return nil, errors.New("quantity must be positive")
		}
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	var created *domain.Order
	err = s.tx.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		unique := make(map[string]struct{})
		ids := make([]string, 0, len(items))
		for _, item := range items {
			if _, ok := unique[item.ProductID]; !ok {
				unique[item.ProductID] = struct{}{}
				ids = append(ids, item.ProductID)
			}
		}
		products, err := s.products.GetByIDsForUpdate(ctx, tx, ids)
		if err != nil {
			return err
		}
		if len(products) != len(ids) {
			return errors.New("product not found")
		}
		productMap := make(map[string]domain.Product, len(products))
		for _, p := range products {
			productMap[p.ID] = p
		}
		total := decimal.Zero
		orderItems := make([]domain.OrderItem, 0, len(items))
		for _, item := range items {
			product, ok := productMap[item.ProductID]
			if !ok {
				return errors.New("product not found")
			}
			if product.Quantity < item.Quantity {
				return errors.New("insufficient stock")
			}
			if err := s.products.UpdateQuantity(ctx, tx, product.ID, -item.Quantity); err != nil {
				return err
			}
			linePrice := product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			total = total.Add(linePrice)
			orderItems = append(orderItems, domain.OrderItem{
				ProductID: product.ID,
				Quantity:  item.Quantity,
				Price:     product.Price,
			})
		}
		order := domain.Order{
			UserID:     userID,
			TotalPrice: total,
			Items:      orderItems,
		}
		created, err = s.orders.CreateOrder(ctx, tx, &order, orderItems)
		return err
	})
	return created, err
}
