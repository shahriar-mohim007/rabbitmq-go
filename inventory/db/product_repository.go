package db

import "inventory/model"

func SaveProduct(product model.Product) error {
	result := DB.Create(&product)
	return result.Error
}
