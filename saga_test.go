package main

import (
	"testing"
	"time"
)

func TestSaga(t *testing.T) {
	t.Run("order should be accepted", func(t *testing.T) {
		wsPublishingChannel := make(chan any)
		osPublishingChannel := make(chan any)

		ws := NewWalletService(osPublishingChannel, wsPublishingChannel)
		os := NewOrderService(wsPublishingChannel, osPublishingChannel)

		ws.users[1] = 100
		orderID := os.New(Order{
			state:  "",
			id:     0,
			userID: 1,
			price:  50,
		})

		time.Sleep(100 * time.Millisecond)
		assert(t, int64(50), ws.users[1])
		assert(t, int64(0), ws.usersLocked[1])
		assert(t, OrderState_ACCEPTED, os.orders[orderID].state)

	})

}
