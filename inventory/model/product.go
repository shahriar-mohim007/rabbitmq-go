package model

type Product struct {
	ID          uint   `gorm:"primaryKey"`
	Type        string `json:"type"`
	Variant     string `json:"variant"`
	Description string `json:"description"`
}
