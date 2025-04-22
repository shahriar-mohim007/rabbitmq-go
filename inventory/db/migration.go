package db

import (
	"fmt"
	"inventory/model"
)

func Migrate() {
	err := DB.AutoMigrate(&model.Product{})
	if err != nil {
		panic("❌ Failed to migrate database: " + err.Error())
	}
	fmt.Println("✅ Migration completed!")
}
