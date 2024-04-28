package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"route256.ozon.ru/project/cart/internals/clients"
	"route256.ozon.ru/project/cart/internals/service"
	"route256.ozon.ru/project/cart/internals/storage"
	"route256.ozon.ru/project/cart/models"
)

type addProductPostRequest struct {
	Count uint16 `valid:"type(uint16)"`
}

type checkoutPostRequest struct {
	User int64 `json:"user" valid:"type(int64)"`
}

type checkoutResponse struct {
	OrderId int64 `json:"orderID"`
}

type cartResponse struct {
	TotalPrice uint32           `json:"total_price"`
	Items      []models.Product `json:"items"`
}

func NewHandler(cartService *service.CartService) http.Handler {
	router := http.NewServeMux()

	router.Handle("POST /user/{user_id}/cart/{sku_id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saveProductHandler(w, r, cartService)
	}))

	router.Handle("DELETE /user/{user_id}/cart/{sku_id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deleteProductHandler(w, r, cartService)
	}))

	router.Handle("DELETE /user/{user_id}/cart", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deleteCartHandler(w, r, cartService)
	}))

	router.Handle("GET /user/{user_id}/cart", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getCartHandler(w, r, cartService)
	}))

	router.Handle("POST /cart/checkout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkoutHandler(w, r, cartService)
	}))

	return applyMiddlewares(router)
}

func applyMiddlewares(router *http.ServeMux) http.Handler {
	return LoggingMiddleware(router)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request, cartService *service.CartService) {
	postRequest, err := validateCheckoutPostRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, productList, err := cartService.GetCart(r.Context(), postRequest.User)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orderId, err := cartService.LomsProvider.CreateOrder(r.Context(), postRequest.User, productList)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cartService.DeleteCart(r.Context(), postRequest.User)

	if err != nil {
		http.Error(w, "Failed to delete cart: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(checkoutResponse{orderId})

	if err != nil {
		http.Error(w, "Failed to encode json response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveProductHandler(w http.ResponseWriter, r *http.Request, cartService *service.CartService) {
	userId, err := validateUserId(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	skuId, err := validateSkuId(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postRequest, err := validateAddProductPostRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if postRequest.Count == 0 {
		http.Error(w, "product count must not be 0", http.StatusBadRequest)
		return
	}

	err = cartService.SaveProductItem(r.Context(), userId, skuId, postRequest.Count)

	if err != nil {
		switch {
		case errors.Is(err, clients.ErrProductNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrProductOutOfStock):
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteProductHandler(w http.ResponseWriter, r *http.Request, cartService *service.CartService) {
	userId, err := validateUserId(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	skuId, err := validateSkuId(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = cartService.DeleteProductItem(r.Context(), userId, skuId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteCartHandler(w http.ResponseWriter, r *http.Request, cartService *service.CartService) {
	userId, err := validateUserId(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = cartService.DeleteCart(r.Context(), userId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getCartHandler(w http.ResponseWriter, r *http.Request, cartService *service.CartService) {
	userId, err := validateUserId(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	total, productList, err := cartService.GetCart(r.Context(), userId)

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, storage.ErrUserCartEmpty):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := cartResponse{total, productList}
	data, err := json.Marshal(response)

	if err != nil {
		http.Error(w, "Failed to encode json response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
