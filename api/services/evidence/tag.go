package evidence

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;not null"`
}
