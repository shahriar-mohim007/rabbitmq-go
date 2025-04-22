package service

import (
	"encoding/json"
	"inventory/model"
	"inventory/repository"
)

type InventoryService interface {
	HandleProductCreated(message []byte) error
}

type inventoryServiceImpl struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) InventoryService {
	return &inventoryServiceImpl{repo: repo}
}

func (s *inventoryServiceImpl) HandleProductCreated(message []byte) error {
	var product model.Product
	if err := json.Unmarshal(message, &product); err != nil {
		return err
	}
	s.repo.Save(product)
	return nil
}
