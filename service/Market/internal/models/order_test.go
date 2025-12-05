package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder_JSONSerialization(t *testing.T) {
	now := time.Now()
	order := Order{
		ID:            1,
		UserID:        10,
		TotalAmount:   150.50,
		Status:        "pending",
		PaymentMethod: "card",
		PaymentStatus: "pending",
		DeliveryAddr:  "123 Main St",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	data, err := json.Marshal(order)
	require.NoError(t, err)

	var decoded Order
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, order.ID, decoded.ID)
	assert.Equal(t, order.UserID, decoded.UserID)
	assert.Equal(t, order.TotalAmount, decoded.TotalAmount)
	assert.Equal(t, order.Status, decoded.Status)
	assert.Equal(t, order.PaymentMethod, decoded.PaymentMethod)
}

func TestOrderItem_Fields(t *testing.T) {
	now := time.Now()
	item := OrderItem{
		ID:        1,
		OrderID:   10,
		ProductID: 5,
		Quantity:  2,
		Size:      "M",
		Price:     50.00,
		CreatedAt: now,
	}

	assert.Equal(t, 1, item.ID)
	assert.Equal(t, 10, item.OrderID)
	assert.Equal(t, 5, item.ProductID)
	assert.Equal(t, 2, item.Quantity)
	assert.Equal(t, "M", item.Size)
	assert.Equal(t, 50.00, item.Price)
}

func TestOrderItem_CalculateLineTotal(t *testing.T) {
	item := OrderItem{
		Quantity: 3,
		Price:    25.50,
	}

	lineTotal := float64(item.Quantity) * item.Price
	assert.Equal(t, 76.50, lineTotal)
}

func TestOrderWithItems_TotalCalculation(t *testing.T) {
	now := time.Now()
	order := OrderWithItems{
		Order: Order{
			ID:          1,
			UserID:      10,
			TotalAmount: 150.00,
			Status:      "pending",
		},
		Items: []OrderItem{
			{ID: 1, OrderID: 1, ProductID: 5, Quantity: 2, Price: 50.00, CreatedAt: now},
			{ID: 2, OrderID: 1, ProductID: 6, Quantity: 1, Price: 50.00, CreatedAt: now},
		},
	}

	var calculatedTotal float64
	for _, item := range order.Items {
		calculatedTotal += float64(item.Quantity) * item.Price
	}

	assert.Equal(t, order.TotalAmount, calculatedTotal)
	assert.Len(t, order.Items, 2)
}

func TestCreateOrderRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{"payment_method":"card","delivery_address":"123 Main Street"}`

	var req CreateOrderRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "card", req.PaymentMethod)
	assert.Equal(t, "123 Main Street", req.DeliveryAddr)
}

func TestUpdateOrderStatusRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{"status":"shipped"}`

	var req UpdateOrderStatusRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "shipped", req.Status)
}

func TestOrder_StatusValues(t *testing.T) {
	validStatuses := []string{"pending", "confirmed", "processing", "shipped", "delivered", "cancelled"}

	for _, status := range validStatuses {
		order := Order{Status: status}
		assert.NotEmpty(t, order.Status)
	}
}

func TestOrder_PaymentStatusValues(t *testing.T) {
	validPaymentStatuses := []string{"pending", "paid", "failed", "refunded"}

	for _, status := range validPaymentStatuses {
		order := Order{PaymentStatus: status}
		assert.NotEmpty(t, order.PaymentStatus)
	}
}

func TestOrderWithItems_Embedding(t *testing.T) {
	order := OrderWithItems{
		Order: Order{
			ID:            1,
			UserID:        10,
			TotalAmount:   100.00,
			Status:        "pending",
			PaymentMethod: "cash",
		},
		Items: []OrderItem{},
	}

	assert.Equal(t, 1, order.ID)
	assert.Equal(t, 10, order.UserID)
	assert.Equal(t, "pending", order.Status)
	assert.Empty(t, order.Items)
}

func TestCreateOrderRequest_PaymentMethods(t *testing.T) {
	paymentMethods := []string{"card", "cash", "online"}

	for _, method := range paymentMethods {
		req := CreateOrderRequest{
			PaymentMethod: method,
			DeliveryAddr:  "Test Address",
		}
		assert.Equal(t, method, req.PaymentMethod)
	}
}
