package service

import (
	"context"
	"errors"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"route256.ozon.ru/project/cart/internals/clients"
	"route256.ozon.ru/project/cart/models"
	"testing"
)

type inputData struct {
	userId int64
	skuId  int64
	count  uint16
}

func TestCartService_SaveProductItem(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name      string
		inputData inputData
		mock      func(l *LomsProviderMock, p *ProductProviderMock, c *CartStorageMock, i inputData, result clients.ProductInfo, wantErr error)
		result    clients.ProductInfo
		wantErr   error
	}{
		{
			name: "should be error if product is not found",
			inputData: inputData{
				userId: 1,
				skuId:  1,
				count:  1,
			},
			mock: func(l *LomsProviderMock, p *ProductProviderMock, c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				p.GetProductMock.Expect(minimock.AnyContext, input.skuId).Return(wantResult, wantErr)
				l.GetStockInfoMock.Expect(minimock.AnyContext, input.skuId).Return(10, nil)
			},
			wantErr: errors.New("product is not found"),
		},
		{
			name: "should be error if saving in storage is failed",
			inputData: inputData{
				userId: 1,
				skuId:  1,
				count:  1,
			},
			mock: func(l *LomsProviderMock, p *ProductProviderMock, c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				productInfoMock := clients.ProductInfo{
					Name:  "Product Name",
					Price: 100,
				}
				p.GetProductMock.Expect(minimock.AnyContext, input.skuId).Return(productInfoMock, nil)
				l.GetStockInfoMock.Expect(minimock.AnyContext, input.skuId).Return(10, nil)
				c.AddItemMock.Expect(minimock.AnyContext, input.skuId, input.userId, productInfoMock.Name, input.count).Return(wantErr)
			},
			wantErr: errors.New("couldn't save in storage"),
		},
		{
			name: "should be saved successfully",
			inputData: inputData{
				userId: 1,
				skuId:  1,
				count:  1,
			},
			mock: func(l *LomsProviderMock, p *ProductProviderMock, c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				productInfoMock := clients.ProductInfo{
					Name:  "Product Name",
					Price: 100,
				}
				p.GetProductMock.Expect(minimock.AnyContext, input.skuId).Return(productInfoMock, wantErr)
				l.GetStockInfoMock.Expect(minimock.AnyContext, input.skuId).Return(10, nil)
				c.AddItemMock.Expect(minimock.AnyContext, input.skuId, input.userId, productInfoMock.Name, input.count).Return(wantErr)
			},
		},
		{
			name: "should be err if product out of stock",
			inputData: inputData{
				userId: 1,
				skuId:  1,
				count:  10,
			},
			mock: func(l *LomsProviderMock, p *ProductProviderMock, c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				productInfoMock := clients.ProductInfo{
					Name:  "Product Name",
					Price: 100,
				}
				p.GetProductMock.Expect(minimock.AnyContext, input.skuId).Return(productInfoMock, nil)
				l.GetStockInfoMock.Expect(minimock.AnyContext, input.skuId).Return(2, nil)
			},
			wantErr: ErrProductOutOfStock,
		},
		{
			name: "should be err if stock service is unavailable",
			inputData: inputData{
				userId: 1,
				skuId:  1,
				count:  1,
			},
			mock: func(l *LomsProviderMock, p *ProductProviderMock, c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				productInfoMock := clients.ProductInfo{
					Name:  "Product Name",
					Price: 100,
				}
				p.GetProductMock.Expect(minimock.AnyContext, input.skuId).Return(productInfoMock, nil)
				l.GetStockInfoMock.Expect(minimock.AnyContext, input.skuId).Return(2, wantErr)
			},
			wantErr: errors.New("stock service is failed"),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			ctx := context.Background()
			productProviderMock := NewProductProviderMock(mc)
			lomsProviderMock := NewLomsProviderMock(mc)
			cartStorageMock := NewCartStorageMock(mc)
			cartService := NewCartService(cartStorageMock, productProviderMock, lomsProviderMock)

			test.mock(lomsProviderMock, productProviderMock, cartStorageMock, test.inputData, test.result, test.wantErr)

			err := cartService.SaveProductItem(ctx, test.inputData.userId, test.inputData.skuId, test.inputData.count)

			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestCartService_DeleteCart(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name      string
		inputData inputData
		mock      func(c *CartStorageMock, i inputData, result clients.ProductInfo, wantErr error)
		result    clients.ProductInfo
		wantErr   error
	}{
		{
			name: "should be deleted successfully",
			inputData: inputData{
				userId: 1,
			},
			mock: func(c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				c.DeleteItemsByUserIdMock.Expect(minimock.AnyContext, input.userId).Return(wantErr)
			},
		},
		{
			name: "should be error if user is not found",
			inputData: inputData{
				userId: 1,
			},
			mock: func(c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				c.DeleteItemsByUserIdMock.Expect(minimock.AnyContext, input.userId).Return(wantErr)
			},
			wantErr: errors.New("failed to delete"),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			ctx := context.Background()
			productProviderMock := NewProductProviderMock(mc)
			lomsProviderMock := NewLomsProviderMock(mc)
			cartStorageMock := NewCartStorageMock(mc)

			cartService := NewCartService(cartStorageMock, productProviderMock, lomsProviderMock)

			test.mock(cartStorageMock, test.inputData, test.result, test.wantErr)

			err := cartService.DeleteCart(ctx, test.inputData.userId)

			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestCartService_DeleteProductItem(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name      string
		inputData inputData
		mock      func(p *ProductProviderMock, c *CartStorageMock, i inputData, result clients.ProductInfo, wantErr error)
		result    clients.ProductInfo
		wantErr   error
	}{
		{
			name: "should be error",
			inputData: inputData{
				userId: 1,
				skuId:  1,
			},
			mock: func(p *ProductProviderMock, c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				c.RemoveItemMock.Expect(minimock.AnyContext, input.userId, input.skuId).Return(wantErr)
			},
			wantErr: errors.New("failed to remove item"),
		},
		{
			name: "should be removed successfully",
			inputData: inputData{
				userId: 1,
				skuId:  1,
			},
			mock: func(p *ProductProviderMock, c *CartStorageMock, input inputData, wantResult clients.ProductInfo, wantErr error) {
				c.RemoveItemMock.Expect(minimock.AnyContext, input.userId, input.skuId).Return(wantErr)
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			ctx := context.Background()
			productProviderMock := NewProductProviderMock(mc)
			lomsProviderMock := NewLomsProviderMock(mc)
			cartStorageMock := NewCartStorageMock(mc)

			cartService := NewCartService(cartStorageMock, productProviderMock, lomsProviderMock)

			test.mock(productProviderMock, cartStorageMock, test.inputData, test.result, test.wantErr)

			err := cartService.DeleteProductItem(ctx, test.inputData.userId, test.inputData.skuId)

			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestCartService_GetCart(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name         string
		inputData    inputData
		mock         func(p *ProductProviderMock, c *CartStorageMock, i inputData, wantTotal uint32, wantProducts []models.Product, wantErr error)
		wantTotal    uint32
		wantProducts []models.Product
		wantErr      error
	}{
		{
			name: "should be calculated successfully",
			inputData: inputData{
				userId: 1,
			},
			mock: func(p *ProductProviderMock, c *CartStorageMock, input inputData, wantTotal uint32, wantProducts []models.Product, wantErr error) {
				productsMap := map[models.Product]uint16{
					wantProducts[1]: wantProducts[1].Count,
					wantProducts[0]: wantProducts[0].Count,
				}
				p.GetProductMock.When(minimock.AnyContext, wantProducts[1].SkuId).Then(
					clients.ProductInfo{Name: wantProducts[1].Name, Price: wantProducts[1].Price},
					wantErr,
				)
				p.GetProductMock.When(minimock.AnyContext, wantProducts[0].SkuId).Then(
					clients.ProductInfo{Name: wantProducts[0].Name, Price: wantProducts[0].Price},
					wantErr,
				)
				c.GetItemsByUserIdMock.Expect(minimock.AnyContext, input.userId).Return(productsMap, nil)
			},
			wantTotal: 700,
			wantProducts: []models.Product{
				{
					SkuId: 1,
					Name:  "Product name 1",
					Count: 5,
					Price: 100,
				},
				{
					SkuId: 2,
					Name:  "Product name 2",
					Count: 1,
					Price: 200,
				},
			},
		},
		{
			name: "should be failed if user is not found",
			inputData: inputData{
				userId: 1,
			},
			mock: func(p *ProductProviderMock, c *CartStorageMock, input inputData, wantTotal uint32, wantProducts []models.Product, wantErr error) {
				c.GetItemsByUserIdMock.Expect(minimock.AnyContext, input.userId).Return(map[models.Product]uint16{}, wantErr)
			},
			wantTotal:    0,
			wantProducts: []models.Product{},
			wantErr:      errors.New("user is not found"),
		},
		{
			name: "should be failed if no products are found",
			inputData: inputData{
				userId: 1,
			},
			mock: func(p *ProductProviderMock, c *CartStorageMock, input inputData, wantTotal uint32, wantProducts []models.Product, wantErr error) {
				productsMap := map[models.Product]uint16{
					{SkuId: 1, Name: "Product name", Count: 5, Price: 100}: 0,
				}

				c.GetItemsByUserIdMock.Expect(minimock.AnyContext, input.userId).Return(productsMap, nil)
				p.GetProductMock.Expect(minimock.AnyContext, 1).Return(clients.ProductInfo{}, wantErr)
			},
			wantTotal:    0,
			wantProducts: []models.Product{},
			wantErr:      errors.New("cart is empty"),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			ctx := context.Background()
			productProviderMock := NewProductProviderMock(mc)
			lomsProviderMock := NewLomsProviderMock(mc)
			cartStorageMock := NewCartStorageMock(mc)

			cartService := NewCartService(cartStorageMock, productProviderMock, lomsProviderMock)

			test.mock(productProviderMock, cartStorageMock, test.inputData, test.wantTotal, test.wantProducts, test.wantErr)

			total, products, err := cartService.GetCart(ctx, test.inputData.userId)

			require.Equal(t, test.wantTotal, total)
			require.Equal(t, test.wantProducts, products)
			require.ErrorIs(t, test.wantErr, err)
		})
	}
}
