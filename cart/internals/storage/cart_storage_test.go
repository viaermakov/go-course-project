package storage

import (
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"route256.ozon.ru/project/cart/models"
	"testing"
)

type inputData struct {
	userId int64
	skuId  int64
	name   string
	count  uint16
}

func TestInMemoryCartStorage_AddItem(t *testing.T) {
	t.Parallel()

	t.Run("should be added successfully", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()

		inputData := inputData{
			userId: 1,
			skuId:  1,
			count:  1,
			name:   "Product name",
		}
		product := models.Product{
			SkuId: inputData.skuId,
			Name:  "Product name",
			Count: 0,
			Price: 0,
		}
		want := map[models.Product]uint16{product: 1}

		err := cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
		require.NoError(t, err)

		got, err := cartStorage.GetItemsByUserId(minimock.AnyContext, inputData.userId)
		require.Equal(t, want, got)
	})

	t.Run("should be summarize with the product in the cart", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()

		inputData := inputData{
			userId: 1,
			skuId:  1,
			count:  2,
			name:   "Product name",
		}

		product := models.Product{
			SkuId: inputData.skuId,
			Name:  "Product name",
			Count: 0,
			Price: 0,
		}
		want := map[models.Product]uint16{product: 4}

		err := cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
		require.NoError(t, err)

		err = cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
		require.NoError(t, err)

		expected, err := cartStorage.GetItemsByUserId(minimock.AnyContext, inputData.userId)
		require.Equal(t, want, expected)
	})
}

func TestInMemoryCartStorage_RemoveItem(t *testing.T) {
	t.Parallel()

	t.Run("should be error trying to remove item without user", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()
		inputData := inputData{
			userId: 1,
			skuId:  1,
			count:  1,
			name:   "Product name",
		}

		err := cartStorage.RemoveItem(minimock.AnyContext, inputData.userId, inputData.skuId)

		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("should be removed successfully", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()
		inputData := inputData{
			userId: 1,
			skuId:  1,
			count:  1,
			name:   "Product name",
		}

		err := cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
		require.NoError(t, err)

		err = cartStorage.RemoveItem(minimock.AnyContext, inputData.userId, inputData.skuId)
		require.NoError(t, err)
	})
}

func TestInMemoryCartStorage_GetItemsByUserId(t *testing.T) {
	t.Run("should be error if user is not found", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()

		inputData := inputData{
			userId: 1,
			skuId:  1,
			count:  1,
			name:   "Product name",
		}

		_, err := cartStorage.GetItemsByUserId(minimock.AnyContext, inputData.userId)
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("should be error if cart is empty", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()

		inputData := inputData{
			userId: 1,
			skuId:  1,
			count:  1,
			name:   "Product name",
		}

		err := cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
		require.NoError(t, err)

		err = cartStorage.RemoveItem(minimock.AnyContext, inputData.userId, inputData.skuId)
		require.NoError(t, err)

		_, err = cartStorage.GetItemsByUserId(minimock.AnyContext, inputData.userId)
		require.ErrorIs(t, err, ErrUserCartEmpty)
	})
}

func TestInMemoryCartStorage_DeleteItemsByUserId(t *testing.T) {
	t.Run("should not be error if user is not found", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()

		inputData := inputData{
			userId: 1,
		}

		err := cartStorage.DeleteItemsByUserId(minimock.AnyContext, inputData.userId)
		require.NoError(t, err)
	})

	t.Run("should be deleted successfully", func(t *testing.T) {
		t.Parallel()

		cartStorage := NewInMemoryCartStorage()

		inputData := inputData{
			userId: 1,
			skuId:  1,
			count:  2,
			name:   "Product name",
		}

		err := cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
		require.NoError(t, err)

		err = cartStorage.DeleteItemsByUserId(minimock.AnyContext, inputData.userId)
		require.NoError(t, err)

		_, err = cartStorage.GetItemsByUserId(minimock.AnyContext, inputData.userId)
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func BenchmarkInMemoryCartStorage_AddItem(b *testing.B) {
	cartStorage := NewInMemoryCartStorage()

	inputData := inputData{
		userId: 1,
		skuId:  1,
		count:  2,
		name:   "Product name",
	}

	for n := 0; n < b.N; n++ {
		_ = cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
	}
}

func BenchmarkInMemoryCartStorage_RemoveItem(b *testing.B) {
	cartStorage := NewInMemoryCartStorage()

	inputData := inputData{
		userId: 1,
		skuId:  1,
		count:  2,
		name:   "Product name",
	}

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		_ = cartStorage.AddItem(minimock.AnyContext, inputData.userId, inputData.skuId, inputData.name, inputData.count)
		b.StartTimer()
		_ = cartStorage.RemoveItem(minimock.AnyContext, inputData.userId, inputData.skuId)
	}
}
