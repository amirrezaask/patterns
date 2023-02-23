package main

import (
	"fmt"
	"sync"
)

/*
	- Order service will recieve a new Order.
	- It creates and Order in PENDING state.
	- Emits new order and it's state.
	- Wallet service receieves the order event and checks the user balance.
	- if user has enough credit it will lock the user credit and emits an event telling that.
	- Finally Order service will finalize the order and emits a new event.
	- Wallet service will transfer the credits.
*/

type OrderID int64
type UserID int64

type OrderState string

type Order struct {
	state  OrderState
	id     OrderID
	userID UserID
	price  int64
}

const (
	OrderState_PENDING  OrderState = "pending"
	OrderState_ACCEPTED OrderState = "accepted"
	OrderState_REJECTED OrderState = "rejected"
)

type OrderService interface {
	New(o Order) OrderID
	GetOrderState(oid OrderID) OrderState
}

type WalletService interface {
	GetUserCredit(uid UserID) int64
	GetUserLockedCredit(uid UserID) int64
}

type WalletEvent struct {
	OrderID int64
	Success bool
}

type orderService struct {
	lock              *sync.RWMutex
	orders            []Order
	listeningChannel  chan any
	publishingChannel chan any
}

func NewOrderService(listeningChannel chan any, publishingChannel chan any) *orderService {
	o := &orderService{
		lock:              &sync.RWMutex{},
		listeningChannel:  listeningChannel,
		publishingChannel: publishingChannel,
	}

	go func() {
		for event := range listeningChannel {
			if e, isWalletEvent := event.(WalletEvent); isWalletEvent {
				fmt.Println("order got wallet event")
				fmt.Printf("%+v\n", e)
				o.lock.RLock()
				order := o.orders[e.OrderID]
				o.lock.RUnlock()
				if e.Success {
					order.state = OrderState_ACCEPTED
				} else {
					order.state = OrderState_REJECTED
				}
				o.lock.Lock()
				o.orders[e.OrderID] = order
				o.lock.Unlock()

				o.publishingChannel <- order
			}
		}
	}()

	return o
}

func (o *orderService) New(order Order) OrderID {
	order.id = OrderID(len(o.orders))
	order.state = OrderState_PENDING
	o.orders = append(o.orders, order)
	o.publishingChannel <- order
	return order.id
}

func (o *orderService) GetOrderState(oid OrderID) OrderState {
	return o.orders[oid].state
}

type walletService struct {
	lock              *sync.RWMutex
	listeningChannel  chan any
	publishingChannel chan any
	users             map[UserID]int64 // UserID -> Credit
	usersLocked       map[UserID]int64
}

func NewWalletService(listeningChannel chan any, publishingChannel chan any) *walletService {
	w := &walletService{
		lock:              &sync.RWMutex{},
		listeningChannel:  listeningChannel,
		publishingChannel: publishingChannel,
		users:             make(map[UserID]int64),
		usersLocked:       make(map[UserID]int64),
	}

	go func() {
		for event := range listeningChannel {
			if order, isOrder := event.(Order); isOrder {
				fmt.Println("wallet service got order")
				fmt.Printf("%+v\n", order)
				if order.state == OrderState_PENDING {
					w.lock.RLock()
					userCredit := w.users[order.userID]
					w.lock.RUnlock()
					if userCredit >= order.price {
						w.lock.Lock()
						w.users[order.userID] = userCredit - order.price
						w.usersLocked[order.userID] = order.price
						w.lock.Unlock()
						publishingChannel <- WalletEvent{
							OrderID: int64(order.id),
							Success: true,
						}
					} else {
						publishingChannel <- WalletEvent{
							OrderID: int64(order.id),
							Success: false,
						}
					}
				} else if order.state == OrderState_REJECTED {
					w.lock.Lock()
					w.users[order.userID] = w.users[order.userID] + order.price
					w.usersLocked[order.userID] = w.usersLocked[order.userID] - order.price
					w.lock.Unlock()
				} else if order.state == OrderState_ACCEPTED {
					w.lock.Lock()
					w.usersLocked[order.userID] = w.usersLocked[order.userID] - order.price
					w.lock.Unlock()
				}
			}
		}

	}()

	return w
}
