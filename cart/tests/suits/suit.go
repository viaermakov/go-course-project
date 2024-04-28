package suits

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"route256.ozon.ru/project/cart/config"
	"route256.ozon.ru/project/cart/internals/app"
	"strings"
)

type CartServiceSuit struct {
	suite.Suite
	BaseUrl string
	Server  *http.Server
}

const prefix = "http://"
const address = "localhost:8080"

func (suit *CartServiceSuit) SetupTest() {
	conf := config.NewConfig()
	server := &http.Server{Addr: address, Handler: app.NewCartServer(conf).Handler}

	suit.Server = server
	suit.BaseUrl = prefix + address

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()
}

func (suit *CartServiceSuit) TearDownTest() {
	if err := suit.Server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}

func (suit *CartServiceSuit) TestSaveProductItemIntegration() {
	saveProductRequest, err := createSaveProductRequest(suit.BaseUrl, 2, 773297411, 5)

	suit.Require().NoError(err)

	client := &http.Client{}

	response, err := client.Do(saveProductRequest)

	suit.Require().NoError(err)
	suit.Require().Equal(http.StatusOK, response.StatusCode)

	getCartRequest, err := createGetCartRequest(suit.BaseUrl, 2)
	suit.Require().NoError(err)

	response, err = client.Do(getCartRequest)
	responseBody, err := io.ReadAll(response.Body)

	wantResponse := `{"total_price":11010,"items":[{"sku_id":773297411,"name":"Кроссовки Nike JORDAN","price":2202,"count":5}]}`

	suit.Require().NoError(err)
	suit.Require().Equal(http.StatusOK, response.StatusCode)
	suit.Require().Equal(wantResponse, string(responseBody))
}

func (suit *CartServiceSuit) TestDeleteUserCartIntegration() {
	saveProductRequest, err := createSaveProductRequest(suit.BaseUrl, 2, 773297411, 5)
	client := http.Client{}
	suit.Require().NoError(err)

	response, err := client.Do(saveProductRequest)

	suit.Require().NoError(err)
	suit.Require().Equal(http.StatusOK, response.StatusCode)

	deleteCartRequest, err := createDeleteCartRequest(suit.BaseUrl, 2)
	suit.Require().NoError(err)

	response, err = client.Do(deleteCartRequest)

	suit.Require().NoError(err)
	suit.Require().Equal(http.StatusNoContent, response.StatusCode)

	getCartRequest, err := createGetCartRequest(suit.BaseUrl, 2)
	suit.Require().NoError(err)

	response, err = client.Do(getCartRequest)
	suit.Require().Equal(http.StatusNotFound, response.StatusCode)
}

func createSaveProductRequest(baseUrl string, userId int64, skuId int64, count uint16) (*http.Request, error) {
	type saveProductParamsRequest struct {
		UserId int64
		SkuId  int64
	}
	type saveProductBodyRequest struct {
		Count uint16
	}

	requestQueryParams := saveProductParamsRequest{UserId: userId, SkuId: skuId}
	requestBodyData := saveProductBodyRequest{Count: count}

	requestBodyJson, err := json.Marshal(requestBodyData)

	if err != nil {
		return nil, err
	}

	saveProductRequest := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf(baseUrl+"/user/%v/cart/%v", requestQueryParams.UserId, requestQueryParams.SkuId),
		strings.NewReader(string(requestBodyJson)),
	)
	saveProductRequest.RequestURI = ""

	return saveProductRequest, nil
}

func createGetCartRequest(baseUrl string, userId int64) (*http.Request, error) {
	getCartRequest := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(baseUrl+"/user/%v/cart", userId),
		nil,
	)
	getCartRequest.RequestURI = ""

	return getCartRequest, nil
}

func createDeleteCartRequest(baseUrl string, userId int64) (*http.Request, error) {
	deleteCartRequest := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf(baseUrl+"/user/%v/cart", userId),
		nil,
	)
	deleteCartRequest.RequestURI = ""

	return deleteCartRequest, nil
}
