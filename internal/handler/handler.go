package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	echoSwagger "github.com/swaggo/echo-swagger"

	"stockpilot/internal/domain"
	"stockpilot/internal/middleware"
	"stockpilot/internal/service"
	"stockpilot/pkg/gonerve/logging"
	sentrymw "stockpilot/pkg/gonerve/sentry"
)

type Handler struct {
	users    *service.UserService
	products *service.ProductService
	orders   *service.OrderService
}

func New(users *service.UserService, products *service.ProductService, orders *service.OrderService) *Handler {
	return &Handler{users: users, products: products, orders: orders}
}

func (h *Handler) Register(e *echo.Echo) {
	g := e.Group("/api/v1")
	g.POST("/users/register", h.RegisterUser)
	g.POST("/products", h.CreateProduct)
	g.GET("/products/:id", h.GetProduct)
	g.POST("/orders", h.CreateOrder)
}

type Server struct {
	echo   *echo.Echo
	addr   string
	server *http.Server
}

func NewServer(addr string, users *service.UserService, products *service.ProductService, orders *service.OrderService, logRequests bool, useSentry bool) (*Server, error) {
	e := echo.New()
	e.HideBanner = true
	if logRequests {
		e.Use(middleware.RequestLogger(logging.GlobalLogger()))
	}
	e.Use(sentrymw.PanicEchoMiddleware)
	if useSentry {
		e.Use(sentrymw.ErrEchoMiddleware)
	}

	h := New(users, products, orders)
	h.Register(e)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	s := &Server{
		echo: e,
		addr: addr,
		server: &http.Server{
			Addr:    addr,
			Handler: e,
		},
	}
	return s, nil
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

type RegisterUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Age       int    `json:"age"`
	IsMarried bool   `json:"is_married"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	FullName  string    `json:"full_name"`
	Age       int       `json:"age"`
	IsMarried bool      `json:"is_married"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterUser godoc
// @Summary Register user
// @Tags users
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "register"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users/register [post]
func (h *Handler) RegisterUser(c echo.Context) error {
	var req RegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request"})
	}
	user, err := h.users.Register(c.Request().Context(), service.RegisterInput{
		Email:     strings.TrimSpace(req.Email),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Password:  req.Password,
		Age:       req.Age,
		IsMarried: req.IsMarried,
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, toUserResponse(user))
}

type CreateProductRequest struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Quantity    int      `json:"quantity"`
	Price       string   `json:"price"`
}

type ProductResponse struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Quantity    int       `json:"quantity"`
	Price       string    `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProduct godoc
// @Summary Create product
// @Tags products
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "create product"
// @Success 201 {object} ProductResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/products [post]
func (h *Handler) CreateProduct(c echo.Context) error {
	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request"})
	}
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid price"})
	}
	product, err := h.products.Create(c.Request().Context(), service.CreateProductInput{
		Description: strings.TrimSpace(req.Description),
		Tags:        req.Tags,
		Quantity:    req.Quantity,
		Price:       price,
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, toProductResponse(product))
}

// GetProduct godoc
// @Summary Get product by id
// @Tags products
// @Produce json
// @Param id path string true "product id"
// @Success 200 {object} ProductResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [get]
func (h *Handler) GetProduct(c echo.Context) error {
	id := c.Param("id")
	product, err := h.products.GetByID(c.Request().Context(), id)
	if err != nil {
		return h.writeError(c, err)
	}
	if product == nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "product not found"})
	}
	return c.JSON(http.StatusOK, toProductResponse(product))
}

type CreateOrderRequest struct {
	UserID string                `json:"user_id"`
	Items  []CreateOrderItemBody `json:"items"`
}

type CreateOrderItemBody struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderResponse struct {
	ID         string              `json:"id"`
	UserID     string              `json:"user_id"`
	CreatedAt  time.Time           `json:"created_at"`
	TotalPrice string              `json:"total_price"`
	Items      []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     string `json:"price"`
}

// CreateOrder godoc
// @Summary Create order
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "create order"
// @Success 201 {object} OrderResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/orders [post]
func (h *Handler) CreateOrder(c echo.Context) error {
	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request"})
	}
	items := make([]service.OrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.OrderItemInput{
			ProductID: strings.TrimSpace(item.ProductID),
			Quantity:  item.Quantity,
		})
	}
	order, err := h.orders.Create(c.Request().Context(), strings.TrimSpace(req.UserID), items)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, toOrderResponse(order))
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func (h *Handler) writeError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}
	status := http.StatusInternalServerError
	switch err.Error() {
	case "user must be at least 18",
		"password must be at least 8 characters",
		"email is required",
		"description is required",
		"quantity cannot be negative",
		"price must be positive",
		"id is required",
		"user id is required",
		"order items are required",
		"product id is required",
		"quantity must be positive",
		"user already exists":
		status = http.StatusBadRequest
	case "user not found", "product not found":
		status = http.StatusNotFound
	case "insufficient stock":
		status = http.StatusConflict
	default:
		status = http.StatusInternalServerError
	}
	return c.JSON(status, ErrorResponse{Message: err.Error()})
}

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		FullName:  u.FullName(),
		Age:       u.Age,
		IsMarried: u.IsMarried,
		CreatedAt: u.CreatedAt,
	}
}

func toProductResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		Description: p.Description,
		Tags:        p.Tags,
		Quantity:    p.Quantity,
		Price:       p.Price.StringFixed(2),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toOrderResponse(o *domain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price.StringFixed(2),
		})
	}
	return OrderResponse{
		ID:         o.ID,
		UserID:     o.UserID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice.StringFixed(2),
		Items:      items,
	}
}
