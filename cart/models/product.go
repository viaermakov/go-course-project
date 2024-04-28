package models

type Product struct {
	SkuId int64  `json:"sku_id"`
	Name  string `json:"name"`
	Price uint32 `json:"price"`
	Count uint16 `json:"count"`
}
