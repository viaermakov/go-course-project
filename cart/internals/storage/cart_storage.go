package storage

import (
	"context"
	"errors"
	"route256.ozon.ru/project/cart/models"
	"sync"
)

type InMemoryCartStorage struct {
	mx       sync.RWMutex
	carts    map[int64]map[int64]uint16
	products map[int64]models.Product
}

var (
	ErrUserNotFound  = errors.New("user is not found")
	ErrUserCartEmpty = errors.New("products are not found")
)

func NewInMemoryCartStorage() *InMemoryCartStorage {
	return &InMemoryCartStorage{
		sync.RWMutex{},
		make(map[int64]map[int64]uint16),
		make(map[int64]models.Product),
	}
}

func (store *InMemoryCartStorage) AddItem(ctx context.Context, userId int64, productId int64, productName string, count uint16) error {
	store.mx.Lock()
	defer store.mx.Unlock()

	if _, ok := store.products[productId]; !ok {
		store.products[productId] = models.Product{SkuId: productId, Name: productName}
	}

	if store.carts[userId] == nil {
		store.carts[userId] = map[int64]uint16{}
	}

	store.carts[userId][productId] += count
	return nil
}

func (store *InMemoryCartStorage) RemoveItem(ctx context.Context, userId int64, productId int64) error {
	store.mx.Lock()
	defer store.mx.Unlock()

	if store.carts[userId] == nil {
		return ErrUserNotFound
	}

	delete(store.carts[userId], productId)
	return nil
}

func (store *InMemoryCartStorage) DeleteItemsByUserId(ctx context.Context, userId int64) error {
	store.mx.Lock()
	defer store.mx.Unlock()

	if store.carts[userId] == nil {
		return nil
	}

	delete(store.carts, userId)
	return nil
}

func (store *InMemoryCartStorage) GetItemsByUserId(ctx context.Context, userId int64) (map[models.Product]uint16, error) {
	store.mx.RLock()
	defer store.mx.RUnlock()

	if store.carts[userId] == nil {
		return nil, ErrUserNotFound
	}

	products := make(map[models.Product]uint16)

	if len(store.carts[userId]) == 0 {
		return nil, ErrUserCartEmpty
	}

	for productId, count := range store.carts[userId] {
		product := store.products[productId]
		products[product] = count
	}

	return products, nil
}
