package mainspec

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"stockpilot/code/tests"
	"stockpilot/internal/handler"
)

var _ = ReportAfterSuite("custom report", func(report Report) {
	tree := tests.NewTreeReport()
	for _, specReport := range report.SpecReports {
		tree.AddSpecReport(specReport)
	}
	println()
	tree.Print("", "\t")
})

var _ = Describe("Stockpilot API", Ordered, func() {
	var (
		userReq       handler.RegisterUserRequest
		productReq    handler.CreateProductRequest
		createdUser   handler.UserResponse
		createdProd   handler.ProductResponse
		createdOrder  handler.OrderResponse
		orderQuantity = 2
	)

	BeforeAll(func() {
		userReq = handler.RegisterUserRequest{
			Email:     fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()),
			FirstName: "John",
			LastName:  "Tester",
			Password:  "Sup3rPass!",
			Age:       32,
			IsMarried: false,
		}
		productReq = handler.CreateProductRequest{
			Description: "Demo product",
			Tags:        []string{"demo", "ginkgo"},
			Quantity:    3,
			Price:       "12.50",
		}
	})

	Describe("User registration", Ordered, func() {
		It("rejects underage users", func() {
			resp, err := TestSuite.ApiClient.RegisterUser(handler.RegisterUserRequest{
				Email:    fmt.Sprintf("teen-%d@example.com", time.Now().UnixNano()),
				Password: "password",
				Age:      16,
			})
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var errResp handler.ErrorResponse
			Expect(decodeBody(resp, &errResp)).To(Succeed())
			Expect(errResp.Message).To(ContainSubstring("at least 18"))
		})

		It("registers a new user", func() {
			resp, err := TestSuite.ApiClient.RegisterUser(userReq)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			Expect(decodeBody(resp, &createdUser)).To(Succeed())
			Expect(createdUser.Email).To(Equal(userReq.Email))
			Expect(createdUser.FullName).To(Equal(userReq.FirstName + " " + userReq.LastName))
			Expect(createdUser.Age).To(Equal(userReq.Age))
			Expect(createdUser.ID).NotTo(BeEmpty())
		})

		It("prevents duplicate registrations", func() {
			resp, err := TestSuite.ApiClient.RegisterUser(userReq)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

			var errResp handler.ErrorResponse
			Expect(decodeBody(resp, &errResp)).To(Succeed())
			Expect(errResp.Message).To(Equal("user already exists"))
		})
	})

	Describe("Product lifecycle", Ordered, func() {
		It("rejects malformed price", func() {
			resp, err := TestSuite.ApiClient.CreateProduct(handler.CreateProductRequest{
				Description: "Broken price",
				Quantity:    1,
				Price:       "abc",
			})
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var errResp handler.ErrorResponse
			Expect(decodeBody(resp, &errResp)).To(Succeed())
			Expect(errResp.Message).To(ContainSubstring("invalid price"))
		})

		It("creates a product", func() {
			resp, err := TestSuite.ApiClient.CreateProduct(productReq)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(decodeBody(resp, &createdProd)).To(Succeed())
			Expect(createdProd.Description).To(Equal(productReq.Description))
			Expect(createdProd.Tags).To(ConsistOf(productReq.Tags))
			Expect(createdProd.Quantity).To(Equal(productReq.Quantity))
			Expect(createdProd.Price).To(Equal(productReq.Price))
		})

		It("retrieves product by id", func() {
			resp, err := TestSuite.ApiClient.GetProduct(createdProd.ID)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			var fetched handler.ProductResponse
			Expect(decodeBody(resp, &fetched)).To(Succeed())
			Expect(fetched.ID).To(Equal(createdProd.ID))
			Expect(fetched.Price).To(Equal(createdProd.Price))
		})
	})

	Describe("Order creation", Ordered, func() {
		It("fails when user does not exist", func() {
			resp, err := TestSuite.ApiClient.CreateOrder(handler.CreateOrderRequest{
				UserID: "missing-user",
				Items: []handler.CreateOrderItemBody{
					{ProductID: createdProd.ID, Quantity: 1},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			var errResp handler.ErrorResponse
			Expect(decodeBody(resp, &errResp)).To(Succeed())
			Expect(errResp.Message).To(Equal("user not found"))
		})

		It("rejects orders exceeding stock", func() {
			resp, err := TestSuite.ApiClient.CreateOrder(handler.CreateOrderRequest{
				UserID: createdUser.ID,
				Items: []handler.CreateOrderItemBody{
					{ProductID: createdProd.ID, Quantity: 10},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusConflict))
			var errResp handler.ErrorResponse
			Expect(decodeBody(resp, &errResp)).To(Succeed())
			Expect(errResp.Message).To(Equal("insufficient stock"))
		})

		It("creates an order and updates quantity", func() {
			resp, err := TestSuite.ApiClient.CreateOrder(handler.CreateOrderRequest{
				UserID: createdUser.ID,
				Items: []handler.CreateOrderItemBody{
					{ProductID: createdProd.ID, Quantity: orderQuantity},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(decodeBody(resp, &createdOrder)).To(Succeed())
			Expect(createdOrder.UserID).To(Equal(createdUser.ID))
			Expect(createdOrder.Items).To(HaveLen(1))
			Expect(createdOrder.Items[0].ProductID).To(Equal(createdProd.ID))
			Expect(createdOrder.Items[0].Quantity).To(Equal(orderQuantity))
			Expect(createdOrder.TotalPrice).To(Equal("25.00"))

			respCheck, err := TestSuite.ApiClient.GetProduct(createdProd.ID)
			Expect(err).NotTo(HaveOccurred())
			defer respCheck.Body.Close()
			Expect(respCheck.StatusCode).To(Equal(http.StatusOK))
			var updated handler.ProductResponse
			Expect(decodeBody(respCheck, &updated)).To(Succeed())
			Expect(updated.Quantity).To(Equal(productReq.Quantity - orderQuantity))
		})
	})
})

func decodeBody(resp *http.Response, out any) error {
	return json.NewDecoder(resp.Body).Decode(out)
}
