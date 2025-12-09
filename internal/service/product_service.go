package service

import (
	"context"

	"github.com/shopspring/decimal"

	"stockpilot/internal/domain"
	"stockpilot/pkg/gonerve/errors"
)

type CreateProductInput struct {
	Description string
	Tags        []string
	Quantity    int
	Price       decimal.Decimal
}

type ProductService struct {
	products domain.ProductRepository
}

func NewProductService(products domain.ProductRepository) *ProductService {
	return &ProductService{products: products}
}

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (*domain.Product, error) {
	if input.Description == "" {
		return nil, errors.New("description is required")
	}
	if input.Quantity < 0 {
		return nil, errors.New("quantity cannot be negative")
	}
	if input.Price.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("price must be positive")
	}
	product := domain.Product{
		Description: input.Description,
		Tags:        input.Tags,
		Quantity:    input.Quantity,
		Price:       input.Price,
	}
	created, err := s.products.CreateProduct(ctx, &product)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.products.GetProductByID(ctx, id)
}
