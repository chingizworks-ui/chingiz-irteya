package concurrency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"stockpilot/internal/handler"
)

var _ = Describe("Concurrent order processing", Ordered, func() {
	const (
		productQuantity = 3
		totalRequests   = 5
	)

	var (
		user    handler.UserResponse
		product handler.ProductResponse
	)

	BeforeAll(func() {
		resp, err := TestSuite.ApiClient.RegisterUser(handler.RegisterUserRequest{
			Email:     fmt.Sprintf("parallel-%d@example.com", time.Now().UnixNano()),
			FirstName: "Parallel",
			LastName:  "Runner",
			Password:  "StrongPassword",
			Age:       30,
		})
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		Expect(decodeBody(resp, &user)).To(Succeed())

		resp, err = TestSuite.ApiClient.CreateProduct(handler.CreateProductRequest{
			Description: "Concurrent product",
			Tags:        []string{"load"},
			Quantity:    productQuantity,
			Price:       "10.00",
		})
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		Expect(decodeBody(resp, &product)).To(Succeed())
	})

	It("does not oversell stock under parallel load", func() {
		var successCount atomic.Int32
		var conflictCount atomic.Int32

		var wg sync.WaitGroup
		wg.Add(totalRequests)

		for i := 0; i < totalRequests; i++ {
			go func(idx int) {
				defer GinkgoRecover()
				defer wg.Done()
				resp, err := TestSuite.ApiClient.CreateOrder(handler.CreateOrderRequest{
					UserID: user.ID,
					Items: []handler.CreateOrderItemBody{
						{ProductID: product.ID, Quantity: 1},
					},
				})
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				switch resp.StatusCode {
				case http.StatusCreated:
					successCount.Add(1)
				case http.StatusConflict:
					conflictCount.Add(1)
				default:
					Fail(fmt.Sprintf("unexpected status code %d at request %d", resp.StatusCode, idx))
				}
			}(i)
		}
		wg.Wait()

		Expect(int(successCount.Load())).To(Equal(productQuantity))
		Expect(int(conflictCount.Load())).To(Equal(totalRequests - productQuantity))

		respCheck, err := TestSuite.ApiClient.GetProduct(product.ID)
		Expect(err).NotTo(HaveOccurred())
		defer respCheck.Body.Close()
		Expect(respCheck.StatusCode).To(Equal(http.StatusOK))
		var after handler.ProductResponse
		Expect(decodeBody(respCheck, &after)).To(Succeed())
		Expect(after.Quantity).To(BeZero())
	})
})

func decodeBody(resp *http.Response, out any) error {
	return json.NewDecoder(resp.Body).Decode(out)
}
