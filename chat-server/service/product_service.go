package service

import (
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
)

func GetAllProducts() (*model.ProductVO, error) {
	var products = make([]model.Product, 0)
	res := repo.QueryAllProduct(&products)
	if res.Error != nil {
		return nil, res.Error
	}
	// 创建一个ProductVO实例
	var productVO model.ProductVO

	// 遍历products切片
	for _, product := range products {
		// 创建一个新的结构体并赋值
		p := struct {
			ID          uint
			Price       float64
			ProductName string
		}{
			ID:          product.ID,
			Price:       product.Price,
			ProductName: product.ProductName,
		}
		// 把新的结构体追加到ProductVO的Products切片中
		productVO.Products = append(productVO.Products, p)
	}
	return &productVO, nil
}
