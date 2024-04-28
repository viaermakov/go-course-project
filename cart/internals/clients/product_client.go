package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ProductClient struct {
	client *http.Client
	cacher Cacher
}

type Cacher interface {
	Add(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Lock()
	Unlock()
}

type ProductInfo struct {
	Name  string `json:"name" redis:"name"`
	Price uint32 `json:"price" redis:"price"`
}

type productRequest struct {
	Token string `json:"token"`
	SkuId uint32 `json:"sku"`
}

const (
	token         = "testtoken"
	baseUrl       = "http://route256.pavl.uk:8080"
	getProductUrl = baseUrl + "/get_product"
)

var (
	ErrProductNotFound = errors.New("product is not found")
)

func NewProductClient(cacher Cacher) *ProductClient {
	return &ProductClient{
		client: &http.Client{
			Transport: &retryRoundTripper{proxy: http.DefaultTransport},
		},
		cacher: cacher,
	}
}

type retryRoundTripper struct {
	proxy http.RoundTripper
}

func (p *ProductClient) GetProduct(ctx context.Context, skuId int64) (ProductInfo, error) {
	var product ProductInfo
	var buf bytes.Buffer

	p.cacher.Lock()
	defer p.cacher.Unlock()

	if cachedResult := p.getCachedResult(ctx, skuId); cachedResult != nil {
		return *cachedResult, nil
	}

	err := json.NewEncoder(&buf).Encode(productRequest{token, uint32(skuId)})

	if err != nil {
		return product, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, getProductUrl, &buf)

	if err != nil {
		return product, err
	}

	resp, err := p.client.Do(req)

	if err != nil {
		return product, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return product, ErrProductNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return product, errors.New("failed to get product info: status code is not ok")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return product, err
	}

	var productInfo ProductInfo

	err = json.Unmarshal(body, &productInfo)

	if err != nil {
		return product, err
	}

	err = p.cacher.Add(ctx, strconv.Itoa(int(skuId)), body)

	if err != nil {
		return ProductInfo{}, err
	}

	return productInfo, err
}

func (p *ProductClient) getCachedResult(ctx context.Context, skuId int64) *ProductInfo {
	var product ProductInfo
	cachedResult, err := p.cacher.Get(ctx, strconv.Itoa(int(skuId)))

	if err == nil && cachedResult != nil {
		err := json.Unmarshal(cachedResult, &product)

		if err != nil {
			log.Println("[cache] get error while parsing", err)
		}

		return &product
	}

	if err != nil {
		log.Println("[cache] get error", err)
	}

	return nil
}

func (roundTripper *retryRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	buf, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	return retry(roundTripper, r, buf, 0)
}

func retry(roundTripper *retryRoundTripper, request *http.Request, rawBody []byte, count int) (*http.Response, error) {
	if count == 3 {
		return nil, errors.New("product service is unavailable")
	}

	request.Body = io.NopCloser(bytes.NewReader(rawBody))
	request.ContentLength = int64(len(rawBody))

	response, err := roundTripper.proxy.RoundTrip(request)

	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode == 420 {
		response.Body.Close()
		return retry(roundTripper, request, rawBody, count+1)
	}

	return response, nil
}
