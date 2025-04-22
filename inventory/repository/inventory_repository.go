package repository

import (
	"inventory/db"
	"inventory/model"
)

type InventoryRepository interface {
	Save(product model.Product)
}

type inventoryRepositoryImpl struct{}

func NewInventoryRepository() InventoryRepository {
	return &inventoryRepositoryImpl{}
}

func (r *inventoryRepositoryImpl) Save(product model.Product) {
	db.SaveProduct(product)
}
