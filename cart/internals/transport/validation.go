package transport

import (
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"io"
	"net/http"
	"strconv"
)

func validateUserId(r *http.Request) (int64, error) {
	ok := govalidator.IsInt(r.PathValue("user_id"))

	if !ok {
		return 0, errors.New("user id is not valid")
	}

	userId, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)

	if err != nil {
		return 0, errors.New("user id is not valid")
	}

	return userId, nil
}

func validateSkuId(r *http.Request) (int64, error) {
	ok := govalidator.IsInt(r.PathValue("sku_id"))

	if !ok {
		return 0, errors.New("product item id is not valid")
	}

	skuId, err := strconv.ParseInt(r.PathValue("sku_id"), 10, 64)

	if err != nil {
		return 0, errors.New("product item id is not valid")
	}

	return skuId, nil
}

func validateAddProductPostRequest(r *http.Request) (addProductPostRequest, error) {
	body, err := io.ReadAll(r.Body)
	postRequest := addProductPostRequest{}

	if err != nil {
		return postRequest, errors.New("body request is not valid")
	}

	err = json.Unmarshal(body, &postRequest)

	if err != nil {
		return postRequest, errors.New("body response isn't a valid json")
	}

	ok, err := govalidator.ValidateStruct(addProductPostRequest{})

	if !ok || err != nil {
		return postRequest, errors.New("body response doesn't correspond required schema")
	}

	return postRequest, nil
}

func validateCheckoutPostRequest(r *http.Request) (checkoutPostRequest, error) {
	body, err := io.ReadAll(r.Body)
	postRequest := checkoutPostRequest{}

	if err != nil {
		return postRequest, errors.New("body request is not valid")
	}

	err = json.Unmarshal(body, &postRequest)

	if err != nil {
		return postRequest, errors.New("body response isn't a valid json")
	}

	ok, err := govalidator.ValidateStruct(checkoutPostRequest{})

	if !ok || err != nil {
		return postRequest, errors.New("body response doesn't correspond required schema")
	}

	return postRequest, nil
}
