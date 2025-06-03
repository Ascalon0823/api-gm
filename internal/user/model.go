package user

import "time"

type User struct {
	ID          uint   `json:"id" bson:"_id,omitempty" gorm:"primaryKey"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email" gorm:"uniqueIndex"`
	Password    string `json:"password"`
	CreatedAt   time.Time
}
