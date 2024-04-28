package service

import (
	"cmp"
	"context"
	"errors"
	"route256.ozon.ru/project/cart/internals/clients"
	"route256.ozon.ru/project/cart/internals/infra/errgrp"
	"route256.ozon.ru/project/cart/internals/infra/limiter"
	"route256.ozon.ru/project/cart/models"
	"slices"
)

type CartService struct {
	LomsProvider    LomsProvider
	ProductProvider ProductProvider
	Store           CartStorage
}

type ProductProvider interface {
	GetProduct(ctx context.Context, skuId int64) (clients.ProductInfo, error)
}

type LomsProvider interface {
	CreateOrder(ctx context.Context, userId int64, items []models.Product) (int64, error)
	GetStockInfo(ctx context.Context, skuId int64) (uint64, error)
}

type CartStorage interface {
	AddItem(ctx context.Context, userId int64, productId int64, productName string, count uint16) error
	RemoveItem(ctx context.Context, userId int64, productId int64) error
	DeleteItemsByUserId(ctx context.Context, userId int64) error
	GetItemsByUserId(ctx context.Context, userId int64) (map[models.Product]uint16, error)
}

var (
	ErrProductOutOfStock = errors.New("product is out of stock")
)

func NewCartService(store CartStorage, productProvider ProductProvider, lomsProvider LomsProvider) *CartService {
	return &CartService{
		lomsProvider,
		productProvider,
		store,
	}
}

func (service *CartService) SaveProductItem(ctx context.Context, userId int64, skuId int64, count uint16) error {
	g, ctxWithCancel := errgrp.WithContext(ctx)

	var product clients.ProductInfo

	g.Go(func() error {
		productInfo, err := service.ProductProvider.GetProduct(ctxWithCancel, skuId)

		if err != nil {
			return err
		}

		product = productInfo
		return nil
	})

	g.Go(func() error {
		availableStocks, err := service.LomsProvider.GetStockInfo(ctxWithCancel, skuId)

		if err != nil {
			return err
		}

		if availableStocks < uint64(count) {
			return ErrProductOutOfStock
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return service.Store.AddItem(ctx, userId, skuId, product.Name, count)
}

func (service *CartService) DeleteCart(ctx context.Context, userId int64) error {
	return service.Store.DeleteItemsByUserId(ctx, userId)
}

func (service *CartService) DeleteProductItem(ctx context.Context, userId int64, skuId int64) error {
	return service.Store.RemoveItem(ctx, userId, skuId)
}

func (service *CartService) GetCart(ctx context.Context, userId int64) (uint32, []models.Product, error) {
	products, err := service.Store.GetItemsByUserId(ctx, userId)

	if err != nil {
		return 0, []models.Product{}, err
	}

	return service.calculateTotal(ctx, products)
}

func (service *CartService) calculateTotal(ctx context.Context, products map[models.Product]uint16) (uint32, []models.Product, error) {
	resultChannel := make(chan models.Product, len(products))

	lim, _ := limiter.NewLimiter(1, 10)
	g, ctxWithCancel := errgrp.WithContext(ctx)

	for product, count := range products {
		lim.Wait()
		skuId := product.SkuId

		g.Go(func() error {
			productInfo, err := service.ProductProvider.GetProduct(ctxWithCancel, skuId)

			if err != nil {
				return err
			}

			resultChannel <- models.Product{
				SkuId: skuId,
				Count: count,
				Name:  productInfo.Name,
				Price: productInfo.Price,
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		close(resultChannel)
		return 0, []models.Product{}, err
	}

	total := uint32(0)
	productsWithPrice := make([]models.Product, len(products))

	for i := 0; i < len(products); i++ {
		result := <-resultChannel
		productsWithPrice[i] = result
		total += result.Price * uint32(result.Count)
	}

	slices.SortFunc(productsWithPrice, func(a, b models.Product) int {
		return cmp.Compare(a.SkuId, b.SkuId)
	})

	return total, productsWithPrice, nil
}
