package main

import (
	"inventory/db"
	"inventory/rabbitmq"
	"inventory/repository"
	"inventory/service"
)

func main() {
	// Step 1: DB connect
	db.ConnectDatabase()

	// Step 2: DB migrate
	db.Migrate()

	repo := repository.NewInventoryRepository()
	svc := service.NewInventoryService(repo)

	rabbitmq.StartConsumer(svc)
}
